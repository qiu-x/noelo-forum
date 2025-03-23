package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrRegister        = errors.New("registration error")
	ErrInvalidUserData = fmt.Errorf("invalid user data: %w", ErrRegister)
	ErrUserExists      = fmt.Errorf("user already exists: %w", ErrRegister)
)

type Storage struct {
	mu sync.Mutex
}

func (s *Storage) AddUser(email, username, pass string) error {
	// Lock this part to avoid races
	s.mu.Lock()
	defer s.mu.Unlock()

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
