package storage

import (
	"errors"
	"fmt"
	"forumapp/tmpl"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type Storage struct {
	mu sync.Mutex
}

var (
	ErrRegister        = errors.New("registration error")
	ErrInvalidUserData = fmt.Errorf("invalid user data: %w", ErrRegister)
	ErrUserExists      = fmt.Errorf("user already exists: %w", ErrRegister)
)

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

func ParseUserResourceURI(path string) (user string, resourceType string, resourcePath string, err error) {
	urlparts := strings.Split(path, "/")
	if len(urlparts) < 2 {
		err = errors.New("invalid URI")
		return
	}
	resourceURI := strings.Split(urlparts[2], ":")
	if len(resourceURI) < 2 {
		err = errors.New("invalid URI")
		return
	}
	user = urlparts[1]
	resourceID := resourceURI[1]
	resourceType = resourceURI[0]

    supportedTypes := []string{ "post", "comment" }
    if !slices.Contains(supportedTypes, resourceType) {
        err = errors.New("unsupported Type:" + resourceType)
		return
    }

	resourcePath, err = url.JoinPath(user, resourceType, resourceID)
	resourcePath = "../storage/users/" + resourcePath
	return
}

func GetPost(resourcePath string, location, user string) (tmpl.TextPost, error) {
	title, err := os.ReadFile(filepath.Join(resourcePath, "title"))
	if err != nil {
		return tmpl.TextPost{}, fmt.Errorf("failed to get post title: %w", err)
	}
	text, err := os.ReadFile(filepath.Join(resourcePath, "text"))
	if err != nil {
		return tmpl.TextPost{}, fmt.Errorf("failed to get post text", err)
	}
	comments, err := getReplies(resourcePath, "comments")
	if err != nil && !os.IsNotExist(err) {
		return tmpl.TextPost{}, fmt.Errorf("failed to get post comments", err)
	}

	return tmpl.TextPost{
		Location: location,
		Title:    string(title),
		Text:     string(text),
		Author:   user,
		Comments: comments,
	}, nil
}

func parseComment(resourcePath string, user string) (tmpl.Comment, bool) {
	location, err := os.ReadFile(filepath.Join(resourcePath, "location"))
	if err != nil {
		return tmpl.Comment{}, false
	}
	text, err := os.ReadFile(filepath.Join(resourcePath, "text"))
	if err != nil {
		return tmpl.Comment{}, false
	}
	replies, err := getReplies(resourcePath, "replies")
	if err != nil {
		return tmpl.Comment{}, false
	}
	return tmpl.Comment{
		Author:   user,
		Location: string(location),
		Text:     string(text),
		Replies:  replies,
	}, true
}

func getReplies(postPath, commentDirName string) ([]tmpl.Comment, error) {
	var comments []tmpl.Comment
	commentDir := filepath.Join(postPath, commentDirName)
	files, err := os.ReadDir(commentDir)
	if err != nil {
		return []tmpl.Comment{}, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		commentURI, err := os.ReadFile(filepath.Join(commentDir, file.Name()))
		if err != nil {
			continue
		}
		user, _, resourcePath, err := ParseUserResourceURI(strings.TrimSpace(string(commentURI)))
		if err != nil {
			continue
		}
		comment, ok := parseComment(resourcePath, user)
		if !ok {
			continue
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

