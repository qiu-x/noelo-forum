package page

import (
	"forumapp/storage"
	"testing"
)

func TestNormalizeResourceIDFromPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path     string
		expected string
	}{
		// Full paths with /u/ prefix
		{"/u/posts/post:123", "posts/post:123"},
		{"/u/users/alice", "users/alice"},
		{"/u/comments/comment:456", "comments/comment:456"},

		// Paths without /u/ prefix
		{"/posts/post:123", "posts/post:123"},
		{"/users/alice", "users/alice"},

		// Bare usernames (for backward compat)
		{"/alice", "alice"},
		{"/u/alice", "alice"},

		// Edge cases
		{"", ""},
		{"/", ""},
		{"/u/", ""},
		{"u/posts/post:123", "posts/post:123"},
		{"posts/post:123", "posts/post:123"},

		// Whitespace handling
		{"  /u/posts/post:123  ", "posts/post:123"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			got := NormalizeResourceIDFromPath(tt.path)
			if got != tt.expected {
				t.Fatalf("normalizeResourceIDFromPath(%q) = %q, expected %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestParsePostID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		uri      string
		expected string
	}{
		{"/u/posts/post:123", "posts/post:123"},
		{"posts/post:123", "posts/post:123"},
		{"/posts/post:123", "posts/post:123"},
		{"/u/posts/post:1735992000000000000", "posts/post:1735992000000000000"},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			t.Parallel()
			got := ParsePostID(tt.uri)
			if got != tt.expected {
				t.Fatalf("parsePostID(%q) = %q, expected %q", tt.uri, got, tt.expected)
			}
		})
	}
}

func TestParseCommentID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		uri      string
		expected string
	}{
		{"/u/comments/comment:456", "comments/comment:456"},
		{"comments/comment:456", "comments/comment:456"},
		{"/comments/comment:456", "comments/comment:456"},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			t.Parallel()
			got := ParseCommentID(tt.uri)
			if got != tt.expected {
				t.Fatalf("parseCommentID(%q) = %q, expected %q", tt.uri, got, tt.expected)
			}
		})
	}
}

func TestPostURLFromIDRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		idStr string
	}{
		{"posts/post:123"},
		{"posts/post:1735992000000000000"},
	}

	for _, tt := range tests {
		t.Run(tt.idStr, func(t *testing.T) {
			t.Parallel()
			// Generate URL from ID
			url := PostURLFromID(storage.PostID(tt.idStr))

			// Expected URL format: /u/{id}
			expected := "/u/" + tt.idStr
			if url != expected {
				t.Fatalf("postURLFromID(%q) = %q, expected %q", tt.idStr, url, expected)
			}

			// Parse back to ID
			parsed := ParsePostID(url)
			if parsed != tt.idStr {
				t.Fatalf("parsePostID(%q) = %q, expected %q", url, parsed, tt.idStr)
			}
		})
	}
}
