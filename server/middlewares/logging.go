package middlewares

import (
	"log"
	"net/http"
	"time"
)

func LoggingMW(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v: %v\n%v", r.Method, r.RequestURI, time.Now().Format(time.RFC850))
		h.ServeHTTP(w, r)
	})
}
