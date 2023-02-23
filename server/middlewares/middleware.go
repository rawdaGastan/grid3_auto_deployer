package middlewares

import (
	"fmt"
	"net/http"
)

func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Method)
		h.ServeHTTP(w, r)
	})
}
