package cloudflare

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRequest(t *testing.T) {
	requireEqual := requireEqual(t)
	r := NewRequest(
		http.MethodGet,
		nil,
		Basic("bob", "secret"),
		Accept(ContentTypeYAML),
		URL(
			Endpoint("https://api.example.com/xx"),
			Path("zones/%d", 99),
			Query("page", "%d", 100),
			Query("per_page", "%d", 20),
			Query("order", "%s", SortAscending),
		),
	)
	requireEqual("api.example.com", r.URL.Host)
	requireEqual("/xx/zones/99", r.URL.Path)
	requireEqual("Basic Ym9iOnNlY3JldA==", r.Header.Get("Authorization"))
	requireEqual("100", r.URL.Query().Get("page"))
}

func TestZones(t *testing.T) {
	requireEqual := requireEqual(t)
	requireEqual(1, 1)
	t.Run("it fetches zones properly", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("wrong Method %q", r.Method)
			}
			if r.Header.Get("Accept") != "application/json" {
				t.Errorf("wrong Accept header %q", r.Header.Get("Accept"))
			}
			if r.Header.Get("Authorization") != "Bearer foo" {
				t.Errorf("wrong Authorization header %q", r.Header.Get("Authorization"))
			}
			if r.URL.Path != "/zones" {
				t.Errorf("wrong URL.Path %q", r.URL.Path)
			}
			switch r.URL.Query().Get("page") {
			case "", "1":
				fmt.Fprint(w, `{"success":true,"result_info":{"page":1,"per_page":2,"count":2,"total_count":3},"result":[{"id":"1","name":"foo.com"},{"id":"2","name":"bar.com"}]}`)
			case "2":
				fmt.Fprint(w, `{"success":true,"result_info":{"page":2,"per_page":2,"count":1,"total_count":2},"result":[{"id":"3","name":"baz.com"}]}`)
			default:
				t.Errorf("wrong page query %q", r.URL.Query().Get("page"))
			}
		}))
		defer ts.Close()
		// url, _ := url.Parse(ts.URL)
		// requireEqual("example.com")
		// zones, err := Zones(URL(url), Token("foo"))
		// if err != nil {
		// 	t.Fatal(err)
		// }
		// requireEqual(3, len(zones))
		// requireEqual("1", zones[0].ID)
		// requireEqual("2", zones[1].ID)
		// requireEqual("3", zones[2].ID)
		// requireEqual("foo.com", zones[0].ID)
		// requireEqual("bar.com", zones[1].ID)
		// requireEqual("baz.com", zones[2].ID)
	})
}

func requireEqual(t *testing.T) (f func(...interface{})) {
	str := func(i interface{}) string {
		return fmt.Sprintf("%v", i)
	}
	f = func(a ...interface{}) {
		t.Helper()
		if len(a) < 2 {
			return
		}
		l, r := str(a[0]), str(a[1])
		if l != r {
			t.Fatalf("%s != %s", l, r)
			return
		}
		f(a[1:])
	}
	return
}
