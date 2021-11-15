package main

import (
	"log"
	"net"
	"net/http"
	"time"
)

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
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
		statusText := http.StatusText(statusCode)
		log.Printf("%d %s", statusCode, statusText)
		http.Error(w, statusText, statusCode)
	}
}
