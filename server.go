package main

import (
	"embed"
)

// public is an embedded file system for the web server.
//go:embed public
var public embed.FS
