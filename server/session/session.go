package session

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const SESSION_COOKIE = "session_token"

var sessions = map[string]session{}

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
	if v, ok := sessions[sessionToken]; ok {
		return v.username, true
	}
	return "", false
}

func Auth(username, pass string) (string, error) {
	// Sanitize username
	username = strings.Replace(username, "/", "∕", -1)

	userdir := "../storage/users/" + username
	storedPass, _ := os.ReadFile(userdir + "/pass")

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
	// Sanitize user data
	username = strings.Replace(username, "/", "∕", -1)
	username = strings.TrimSpace(username)

	if !strings.Contains(email, "@") {
		return ErrInvalidUserData
	}

	if email == "" || pass == "" || username == "" {
		return ErrInvalidUserData
	}

	userdir := "../storage/users/" + username
	if _, err := os.Stat(userdir); !os.IsNotExist(err) {
		return ErrUserExists
	}

	// Create the directory
	err := os.Mkdir(userdir, 0755)
	if err != nil {
		return err
	}

	fmt.Println("User directory created successfully:", userdir)

	f, err := os.Create(userdir + "/email")
	if err != nil {
		return err
	}
	f.WriteString(email)
	f.Close()

	hashed, _ := bcrypt.GenerateFromPassword([]byte(pass), 8)
	f, err = os.Create(userdir + "/pass")
	if err != nil {
		return err
	}
	f.Write(hashed)
	f.Close()

	return nil
}
