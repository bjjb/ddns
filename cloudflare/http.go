package cloudflare

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
)

// A ReqMod function returns a cloned and modified request.
type ReqMod func(*http.Request) *http.Request

// apply runs the functions over the request
func apply(r *http.Request, f ...func(*http.Request) *http.Request) *http.Request {
	for _, f := range f {
		r = f(r)
	}
	return r
}

// Path makes a function which returns a clone of the URL with the path set to
// v (which may be a fmt string, into which a will be interpolated).
func Path(v string, a ...interface{}) func(*url.URL) *url.URL {
	return func(in *url.URL) *url.URL {
		out, err := url.Parse(in.String())
		if err != nil {
			panic(err)
		}
		out.Path = path.Join(out.Path, fmt.Sprintf(v, a...))
		return out
	}
}

// Query makes a function which returns a clone of the URL with the query
// paramater k set to v (which may be a fmt string, into which a will be
// interpolated).
func Query(k, v string, a ...interface{}) func(*url.URL) *url.URL {
	return func(in *url.URL) *url.URL {
		q := in.Query()
		out, err := url.Parse(in.String())
		if err != nil {
			panic(err)
		}
		q.Set(k, fmt.Sprintf(v, a...))
		out.RawQuery = q.Encode()
		return out
	}
}

// A URLFunc returns a copy of a URL with some modification.
type URLFunc func(*url.URL) *url.URL

// URL makes a function which returns a clone of the request with the URL
// changed according to URL modification funcs. If u is nil, the original URL
// is retained before applying the funcs.
func URL(funcs ...func(u *url.URL) *url.URL) ReqMod {
	return func(in *http.Request) *http.Request {
		out := in.Clone(in.Context())
		for _, f := range funcs {
			out.URL = f(out.URL)
		}
		return out
	}
}

// A SortDirection is a valid direction parameter
type SortDirection string

// A ContentType is a valid Content-Type for a CloudFlare response body.
type ContentType string

const (
	// SortAscending means ascending sort order
	SortAscending SortDirection = "ASC"
	// SortDescending means descending sort order
	SortDescending SortDirection = "DESC"
)

const (
	// ContentTypeJSON means JSON.
	ContentTypeJSON ContentType = "application/json"
	// ContentTypeText means plain text.
	ContentTypeText ContentType = "text/plain"
	// ContentTypeYAML means YAML.
	ContentTypeYAML ContentType = "application/x-yaml"
)

// Header creates a function which returns a function which returns a clone of
// the given http.Request with the header k set to v (which may be a fmt
// string, in which case a is interpolated).
func Header(k string) func(v string, a ...interface{}) ReqMod {
	return func(v string, a ...interface{}) ReqMod {
		return func(in *http.Request) *http.Request {
			out := in.Clone(in.Context())
			out.Header.Set(k, fmt.Sprintf(v, a...))
			return out
		}
	}
}

// Authorization returns a function which returns a clone of the given
// http.Request
// with the Authorization header set to v.
var Authorization = Header("Authorization")

// Token creates a function which returns a clone of the request with the
// Authorization header set to "Bearer <token>".
func Token(token string) ReqMod {
	return Authorization("Bearer " + token)
}

// APIKeys creates a function which returns a clone of the request with the
// X-Auth-Email and X-Auth-Key headers set.
func APIKeys(email, key string) ReqMod {
	return func(r *http.Request) *http.Request {
		r = r.Clone(r.Context())
		r.Header.Set("X-Auth-Email", email)
		r.Header.Set("X-Auth-Key", key)
		return r
	}
}

// Accept creates a function which returns a clone of the given http.Request
// with the Accept header set to contentType.
func Accept(contentType ContentType) ReqMod {
	return Header("Accept")("%s", contentType)
}

// Basic creates a function which returns a clone of the given http.Request
// with the basic auth set to username and password.
func Basic(username, password string) ReqMod {
	return func(in *http.Request) *http.Request {
		out := in.Clone(in.Context())
		out.SetBasicAuth(username, password)
		return out
	}
}

// An ErrNotOK is returned when we're expected a 200 but get something else.
type ErrNotOK int

func (e ErrNotOK) Error() string { return http.StatusText(int(e)) }

// An ErrContentType is returned when we cannot parse a response's body
// because it's got the wrong Content-Type header.
type ErrContentType string

func (e ErrContentType) Error() string {
	return fmt.Sprintf("cannot parse %s", string(e))
}
