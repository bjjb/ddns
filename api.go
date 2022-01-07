package main

import "net/http"

type api struct {
	mux *http.ServeMux
}

func (h *api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.mux == nil {
		h.mux = http.NewServeMux()
	}
	h.mux.ServeHTTP(w, r)
}
