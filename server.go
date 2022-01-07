package main

import (
	"embed"
	"io/fs"
	"net/http"
)

// webFS is an embedded file system for the web server.
//go:embed public
var webFS embed.FS

// webDir is used as the webFS, if set.
var webDir = env("DDNS_WEB_DIR", "")

// handler is the request handler.
var handler = func() http.Handler {
	fs, err := fs.Sub(webFS, "public")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(fs))

	mux := http.NewServeMux()
	mux.Handle("/oauth2/", &oauth2{
		"/oauth2/google":  &google{},
		"/oauth2/gitlab":  &gitlab{},
		"/oauth2/github":  &github{},
		"/oauth2/twitter": &twitter{},
	})

	if webDir != "" {
		fileServer = http.FileServer(http.Dir(webDir))
	}

	mux.Handle("/api", &api{})

	mux.Handle("/", fileServer)
	return mux
}
