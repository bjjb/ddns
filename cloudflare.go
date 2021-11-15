package main

import (
	"net/http"
)

var cloudflareHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
})
