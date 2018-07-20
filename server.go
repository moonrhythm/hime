package hime

import (
	"net/http"
	"time"
)

// StartHTTPSRedirectServer starts http to https redirect server
func StartHTTPSRedirectServer(addr string) error {
	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Connection", "close")
			http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
		}),
	}
	return srv.ListenAndServe()
}
