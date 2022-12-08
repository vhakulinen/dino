package httputil_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/vhakulinen/dino/httputil"
)

func ExampleChainMiddleware() {
	h1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("h1\n"))
			next.ServeHTTP(w, r)
		})
	}

	h2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("h2\n"))
			next.ServeHTTP(w, r)
		})
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello!\n"))
	})

	h = httputil.ChainMiddleware(h, h1, h2).ServeHTTP

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	h(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	fmt.Print(string(body))

	// Output:
	// h1
	// h2
	// Hello!
}
