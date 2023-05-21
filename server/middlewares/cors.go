// Package middlewares for middleware between api and backend
package middlewares

import "net/http"

// EnableCors enables cors middleware
func EnableCors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/*
			response writer is an interface. you should never pass a pointer to an interface
			because the value of an interface can itself be a pointer

			instead just pass w.

			otherwise you will have to do weird stuff like (*w) below!

			when you receive an interface you should always pass it around as is, it's up
			to the implementor to decide if what implements the interface is a pass by value
			or pass by pointe. that's because an interface can be either implemented on a value or a pointer
		*/
		setupCorsResponse(w, r)
		h.ServeHTTP(w, r)
	})
}

func setupCorsResponse(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")

	if req.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
}
