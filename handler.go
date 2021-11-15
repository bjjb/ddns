package main

import (
	"fmt"
	"net/http"

	"github.com/bjjb/ddns/cloudflare"
)

var handler = func() http.Handler {
	mux := http.NewServeMux()
	for k, v := range handlers {
		mux.Handle(fmt.Sprintf("/%s/", k), v)
	}
	return mux
}()

var handlers = map[string]http.Handler{
	"cloudflare": &cloudflare.Handler{},
}
