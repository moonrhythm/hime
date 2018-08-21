package hime

import (
	"net/http"
	"time"
)

// HTTPSRedirect type
type HTTPSRedirect struct {
	Addr string `json:"addr"`
}

// Server generates https redirect server
func (s HTTPSRedirect) Server() *http.Server {
	return &http.Server{
		Addr:         s.Addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      &s,
	}
}

func (s *HTTPSRedirect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "close")
	http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
}

// StartHTTPSRedirectServer starts http to https redirect server
func StartHTTPSRedirectServer(addr string) error {
	return HTTPSRedirect{Addr: addr}.Server().ListenAndServe()
}
