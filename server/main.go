package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"forumapp/session"
	"forumapp/storage"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
)

type Config struct {
	Host string
	Port string
}

func newConfig(args []string) (Config, error) {
	var cfg Config
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.StringVar(&cfg.Host, "host", "0.0.0.0", "Application Host IP")
	flags.StringVar(&cfg.Port, "port", "8080", "Application Port")
	err := flags.Parse(args[1:])
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func NewServer(
	logger *slog.Logger,
	ses *session.Sessions,
	store *storage.Store,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(
		mux,
		ses,
		store,
	)
	var handler http.Handler = mux
	handler = LoggerMiddleware(logger, handler)
	return handler
}

func getCouchDBConfigFromEnv() (dsn, dbName string) {
	dsn = os.Getenv("COUCHDB_DSN")
	if dsn == "" {
		dsn = "http://admin:3344@localhost:5984" // dev default
	}
	dbName = os.Getenv("COUCHDB_DB")
	if dbName == "" {
		dbName = "forum"
	}
	return
}

func run(ctx context.Context, w io.Writer, e io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	cfg, err := newConfig(args)
	if err != nil {
		_, _ = fmt.Fprintf(e, "error parsing cmd args: %s\n", err)
	}

	logger := slog.New(tint.NewHandler(w, nil))
	sessions := session.NewSessions()

	// Initialize CouchDB backend from environment
	couchDSN, couchDB := getCouchDBConfigFromEnv()
	db, err := storage.ConnectCouchDB(ctx, couchDSN, couchDB)
	if err != nil {
		_, _ = fmt.Fprintf(e, "failed to connect to CouchDB: %s\n", err)
		return err
	}

	backend := storage.NewCouchBackend(db, logger)
	store := storage.NewStore(storage.WithBackend(backend))

	srv := NewServer(
		logger,
		sessions,
		store,
	)
	httpServer := &http.Server{
		Addr:           net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:        srv,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		logger.Info("Listening on",
			slog.String("host", cfg.Host),
			slog.String("port", cfg.Port),
		)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			_, _ = fmt.Fprintf(e, "error listening and serving: %s\n", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down gracefully...")
	shutdownCtx, timeoutCancel := context.WithTimeout(ctx, 10*time.Second)
	defer timeoutCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		_, _ = fmt.Fprintf(e, "error shutting down http server: %s\n", err)
		return err
	}

	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Stderr, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
