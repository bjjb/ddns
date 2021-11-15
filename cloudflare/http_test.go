package cloudflare

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHeader(t *testing.T) {
	requireEqual := requireEqual(t)
	r, _ := http.NewRequest(http.MethodHead, "http://example.com/", nil)
	requireEqual("-y-", Header("x")("-%s-", "y")(r).Header.Get("x"))
}

func TestToken(t *testing.T) {
	requireEqual := requireEqual(t)
	r, _ := http.NewRequest(http.MethodHead, "http://example.com/", nil)
	requireEqual("Bearer token", Token("token")(r).Header.Get("Authorization"))
}

func TestContentType(t *testing.T) {
	requireEqual := requireEqual(t)
	r, _ := http.NewRequest(http.MethodHead, "http://example.com/", nil)
	requireEqual(ContentTypeText, Accept(ContentTypeText)(r).Header.Get("Accept"))
}

func TestPath(t *testing.T) {
	requireEqual := requireEqual(t)
	r, _ := http.NewRequest(http.MethodHead, "http://example.com/foo", nil)
	requireEqual("/bar", URL(Path("bar"))(r).URL.Path)
}

func TestDo(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer ts.Close()
	if called != called {
	}
	// r, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	// resp, err := Do(r)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if !called {
	// 	log.Fatal("wasn't called")
	// }
}
