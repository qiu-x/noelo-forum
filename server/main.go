package main

import (
	"context"
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
	err := flags.Parse(args[1:])
	flags.StringVar(&cfg.Host, "host", "0.0.0.0", "Application Host IP")
	flags.StringVar(&cfg.Port, "port", "8080", "Application Port")
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func NewServer(
	logger *slog.Logger,
	ses *session.Sessions,
	strg *storage.Storage,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(
		mux,
		ses,
		strg,
	)
	var handler http.Handler = mux
	handler = LoggerMiddleware(logger, handler)
	return handler
}

func run(ctx context.Context, w io.Writer, e io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	cfg, err := newConfig(args)
	if err != nil {
		fmt.Fprintf(e, "error parsing cmd args: %s\n", err)
	}

	strg, err := storage.NewStorage()
	if err != nil {
		fmt.Fprintf(e, "failed to initialize storage: %s\n", err)
	}

	mux := http.NewServeMux()
	logger := slog.New(tint.NewHandler(w, nil))
	sessions := session.NewSessions()

	addRoutes(mux, sessions, strg)
	srv := NewServer(
		logger,
		sessions,
		strg,
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
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(e, "error listening and serving: %s\n", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down gracefully...")
	shutdownCtx := context.Background()
	shutdownCtx, timeoutCancel := context.WithTimeout(shutdownCtx, 10*time.Second)
	defer timeoutCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		fmt.Fprintf(e, "error shutting down http server: %s\n", err)
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
