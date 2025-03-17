package pages

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

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

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		loginAction(w, r)
	} else if r.Method == "GET" {
		loginPage(w, r)
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	isLoggedIn := false
	username := ""

	sessionCookie, err := r.Cookie("session_token")
	if err == nil {
		if v, ok := sessions[sessionCookie.Value]; ok {
			username = v.username
			isLoggedIn = true
		}
	}

	page := struct {
		PageName   string
		Username   string
		IsLoggedIn bool
		Content    []struct {
			Title    string
			Author   string
			PostLink string
		}
	}{
		Username:   username,
		IsLoggedIn: isLoggedIn,
	}

	t := template.Must(template.ParseFiles("../templates/page.template", "../templates/login.template"))
	err = t.Execute(w, page)
	if err != nil {
		log.Println("\"login\" page generation failed")
	}
}

func loginAction(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("uname")
	pass := r.FormValue("psw")
	log.Println("psw:", string(pass))

	// Sanitize username
	username = strings.Replace(username, "/", "∕", -1)

	userdir := "../storage/users/" + username
	storedPass, _ := os.ReadFile(userdir + "/pass")

	err := bcrypt.CompareHashAndPassword(storedPass, []byte(pass))
	if err != nil {
		// TODO: Handle login failure
		log.Println("Bad creds!")
		w.Write([]byte("Bad creds!"))
		return
	}

	sessionToken := generateSessionToken()

	sessions[sessionToken] = session{
		username: username,
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "session_token",
		Value: sessionToken,
	})

	http.Redirect(w, r, "/active", http.StatusSeeOther)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		registerAction(w, r)
	} else if r.Method == "GET" {
		registerPage(w, r, "none")
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
}

func registerPage(w http.ResponseWriter, r *http.Request, status string) {
	isLoggedIn := false
	username := ""

	sessionCookie, err := r.Cookie("session_token")
	if err == nil {
		if v, ok := sessions[sessionCookie.Value]; ok {
			username = v.username
			isLoggedIn = true
		}
	}

	page := struct {
		PageName   string
		Username   string
		IsLoggedIn bool
		Content    struct {
			RegisterStatus string
		}
	}{
		Username:   username,
		IsLoggedIn: isLoggedIn,
		PageName:   "",
		Content:    struct{ RegisterStatus string }{status},
	}

	t := template.Must(template.ParseFiles("../templates/page.template", "../templates/register.template"))
	err = t.Execute(w, page)
	if err != nil {
		log.Println("\"register\" page generation failed")
	}
}

var (
	ErrRegister        = errors.New("registration error")
	ErrInvalidUserData = fmt.Errorf("invalid user data: %w", ErrRegister)
	ErrUserExists      = fmt.Errorf("user already exists: %w", ErrRegister)
)

func registerAction(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	username := r.FormValue("uname")
	pass := r.FormValue("psw")

	err := addUser(email, username, pass)

	if errors.Is(err, ErrInvalidUserData) {
		registerPage(w, r, "invalid")
		return
	} else if errors.Is(err, ErrUserExists) {
		registerPage(w, r, "exists")
		return
	} else if err != nil {
		log.Println("Account creation error:", err)
		registerPage(w, r, "unknown") // should never happen
		return
	}

	registerPage(w, r, "success")
}

func addUser(email, username, pass string) error {
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
