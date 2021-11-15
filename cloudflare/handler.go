package cloudflare

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"path"
)

// A Handler deals with requests and talks to CloudFlare using its client.
type Handler struct {
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpError := httpError(w)
	var client *Client
	// Get the token or the API authorization from the basic auth header.
	user, pass, ok := r.BasicAuth()
	switch {
	case ok && (user == "_token_" || user == ""):
		client = New(Token(pass))
	case ok:
		client = New(APIKeys(user, pass))
	default:
		httpError(http.StatusUnauthorized)
		return
	}
	recordName := r.URL.Path[12:]
	if recordName == "" {
		match, err := path.Match(r.Header.Get("Accept"), "application/json")
		if err != nil {
			log.Print(err)
			httpError(http.StatusNotAcceptable)
		}
		if match {
			zones, err := client.GetZones()
			if err != nil {
				log.Print(err)
				httpError(http.StatusInternalServerError)
				return
			}
			if err := json.NewEncoder(w).Encode(zones); err != nil {
				log.Print(err)
				httpError(http.StatusInternalServerError)
				return
			}
			return
		}
		http.NotFound(w, r)
	}
	recordValue, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Print(err)
		httpError(http.StatusBadRequest)
		return
	}
	recordType := recordType(recordValue)
	log.Printf("%v - setting %s to (%s) %s", client, recordName, recordType, recordValue)
	fmt.Fprintf(w, "%v - setting %s to (%s) %s", client, recordName, recordType, recordValue)
}

func recordType(content string) string {
	ip := net.ParseIP(content)
	switch {
	case ip == nil:
		return "CNAME"
	case ip.To4() == nil:
		return "AAAA"
	default:
		return "A"
	}
}

func httpError(w http.ResponseWriter) func(int) {
	return func(statusCode int) {
		http.Error(w, http.StatusText(statusCode), statusCode)
	}
}
