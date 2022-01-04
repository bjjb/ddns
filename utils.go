package main

import (
	"log"
	"net"
	"net/http"
	"time"
)

// env tries to use `genenv` to obtain a string value `k`, falling back to
// return `fb` if the result is a zero string.
func env(k, fb string) string {
	if v := getenv(k); v != "" {
		return v
	}
	return fb
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}

func detectRecordType(content string) string {
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
		statusText := http.StatusText(statusCode)
		log.Printf("%d %s", statusCode, statusText)
		http.Error(w, statusText, statusCode)
	}
}
