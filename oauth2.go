package main

import (
	"fmt"
	"net/http"
)

// oauth2 is a http.Handler mapping paths to OAuth2 implementations.
type oauth2 map[string]http.Handler

// ServeHTTP implement http.Handler.
func (h oauth2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v, found := h[r.URL.Path]
	switch {
	case found:
		v.ServeHTTP(w, r)
		return
	case r.Method == "GET" && r.URL.Path == "/oauth2/providers":
		fmt.Fprint(w, "providers...")
		return
	default:
		http.NotFound(w, r)
	}
}

// A google is an oauth2 for Google.
type google struct{}

func (h *google) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "google", http.StatusNotImplemented)
}

// A google is an oauth2 for GitLab.
type gitlab struct{}

func (h *gitlab) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "gitlab", http.StatusNotImplemented)
}

// A github is an oauth2 for GitHub.
type github struct{}

func (h *github) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "github", http.StatusNotImplemented)
}

// A twitter is an oauth2 for Twitter.
type twitter struct{}

func (h *twitter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "twitter", http.StatusNotImplemented)
}
