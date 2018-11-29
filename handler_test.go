package cfddns

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func TestMain(t *testing.T) {
	ts := httptest.NewServer(&Handler{})
	defer ts.Close()

	url, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Failed to parse ts URL %s: %e", ts.URL, err)
	}
	url.Path = "/zones"

}
