package session

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const SessionCookie = "session_token"

var (
	sessions = map[string]session{}
	mu       = sync.Mutex{}
)

type session struct {
	username string
}

func generateSessionToken() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var result string
	for range 50 {
		digit := r.Intn(10)
		result += strconv.Itoa(digit)
	}
	return result
}

func CheckAuth(sessionToken string) (string, bool) {
	// Reads from map are thread-safe
	if v, ok := sessions[sessionToken]; ok {
		return v.username, true
	}
	return "", false
}

func Auth(username, pass string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	// Sanitize username
	username = strings.Replace(username, "/", "∕", -1)

	userdir := filepath.Join("../storage/users/", username)
	storedPass, _ := os.ReadFile(filepath.Join(userdir, "/pass"))

	err := bcrypt.CompareHashAndPassword(storedPass, []byte(pass))
	if err != nil {
		return "", err
	}

	sessionToken := generateSessionToken()

	sessions[sessionToken] = session{
		username: username,
	}
	return sessionToken, nil
}

var (
	ErrRegister        = errors.New("registration error")
	ErrInvalidUserData = fmt.Errorf("invalid user data: %w", ErrRegister)
	ErrUserExists      = fmt.Errorf("user already exists: %w", ErrRegister)
)

func AddUser(email, username, pass string) error {
	// Lock this part to avoid races
	mu.Lock()
	defer mu.Unlock()

	// Sanitize user data
	username = strings.Replace(username, "/", "∕", -1)
	username = strings.TrimSpace(username)

	if !strings.Contains(email, "@") {
		return ErrInvalidUserData
	}

	if email == "" || pass == "" || username == "" {
		return ErrInvalidUserData
	}

	userdir := filepath.Join("../storage/users/", username)
	if _, err := os.Stat(userdir); !os.IsNotExist(err) {
		return ErrUserExists
	}

	// Create the directory and sub-directories
	err := os.Mkdir(userdir, 0755)
	if err != nil {
		return err
	}
	err = os.Mkdir(filepath.Join(userdir, "comment"), 0755)
	if err != nil {
		return err
	}
	err = os.Mkdir(filepath.Join(userdir, "post"), 0755)
	if err != nil {
		return err
	}

	fmt.Println("User directory created successfully:", userdir)

	f, err := os.Create(filepath.Join(userdir, "/email"))
	if err != nil {
		return err
	}
	_, _ = f.WriteString(email)
	f.Close()

	hashed, _ := bcrypt.GenerateFromPassword([]byte(pass), 8)
	f, err = os.Create(filepath.Join(userdir, "/pass"))
	if err != nil {
		return err
	}
	_, _ = f.Write(hashed)
	f.Close()

	return nil
}
