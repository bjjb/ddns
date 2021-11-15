package cloudflare

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// DefaultEndpoint is the default CloudFlare API (v4) client URL
const DefaultEndpoint = "https://api.cloudflare.com/client/v4"

var defaultEndpoint = func(r *http.Request) *http.Request {
	u, err := url.Parse(DefaultEndpoint)
	if err != nil {
		panic(err)
	}
	return apply(r, Endpoint(u))
}

// Endpoint makes a function which returns a cloned request with the URL set
// to u.
var Endpoint = func(u *url.URL) func(r *http.Request) *http.Request {
	return func(r *http.Request) *http.Request {
		r = r.Clone(r.Context())
		r.URL = u
		return r
	}
}

// A Zone is a CloudFlare DNS zone
type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// A Record is a CloudFlare DNS record
type Record struct {
	ID       string
	Name     string
	Type     string
	Content  string
	TTL      time.Duration
	ZoneID   string
	ZoneName string
}

// Error messages are sometimes included in CloudFlare response bodies.
type Error struct {
	Code       int      `json:"code"`
	Message    string   `json:"message"`
	ErrorChain []*Error `json:"error_chain"`
}

func (e *Error) String() string {
	return fmt.Sprintf("[%d] %s {%s}", e.Code, e.Message, e.ErrorChain)
}

// metadata is a standard set of fields in a CloudFlare response body.
type metadata struct {
	Success  bool     `json:"status"`
	Errors   []*Error `json:"errors"`
	Messages []string `json:"messages"`
}

func (m *metadata) Error() string {
	return fmt.Sprintf("%v", m.Errors)
}

// A ResultInfo contains pagination info in a CloudFlare response body.
type ResultInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
}

// A Result is a typical CloudFlare response body
type Result struct {
	*metadata
	ResultInfo *ResultInfo `json:"result_info"`
}

// NewRequest returns a new *http.Request with the given method and body, and
// with all funcs applied to it.
func NewRequest(method string, body io.Reader, f ...ReqMod) *http.Request {
	env := func(k string) string { return os.Getenv("CLOUDFLARE_" + k) }
	r, err := http.NewRequest(method, DefaultEndpoint, body)
	if err != nil {
		panic(err)
	}
	r = Accept(ContentTypeJSON)(r) // JSON by default
	if u, p := env("EMAIL"), env("KEY"); u != "" && p != "" {
		r = Basic(u, p)(r)
	}
	if v := os.Getenv("TOKEN"); v != "" {
		r = Token(v)(r)
	}
	for _, mod := range f {
		r = mod(r)
	}
	return r
}
