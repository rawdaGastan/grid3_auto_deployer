// Package middlewares for middleware between api and backend
package middlewares

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/codescalers/cloud4students/models"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

/*
this is totally 100% wrong the way you filter out which paths need to bt included in the admin access

check the other comment on the route registration method.

the Router does this if u using subrouter correctly
*/
// AdminAccess to authorize admins in requests
func AdminAccess(includedRoutes []*mux.Route, db models.DB) func(http.Handler) http.Handler {
	// Cache the regex object of each route (obviously for performance purposes)
	var includedRoutesRegexp []*regexp.Regexp
	rl := len(includedRoutes)
	for i := 0; i < rl; i++ {
		r := includedRoutes[i]
		pathRegexp, _ := r.GetPathRegexp()
		regx, _ := regexp.Compile(pathRegexp)
		includedRoutesRegexp = append(includedRoutesRegexp, regx)
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			include := false
			requestMethod := r.Method
			for i := 0; i < rl; i++ {
				includedRoute := includedRoutes[i]
				methods, _ := includedRoute.GetMethods()
				ml := len(methods)
				methodMatched := false
				if ml < 1 {
					methodMatched = true
				} else {
					for j := 0; j < ml; j++ {
						if methods[j] == requestMethod {
							methodMatched = true
							break
						}
					}
				}
				if methodMatched {
					uri := r.RequestURI
					if includedRoutesRegexp[i].MatchString(uri) {
						include = true
						break
					}
				}
			}
			if include {
				userID := r.Context().Value(UserIDKey("UserID")).(string)
				user, err := db.GetUserByID(userID)
				if err == gorm.ErrRecordNotFound {
					writeErrResponse(r, w, http.StatusNotFound, "User is not found")
					return
				}
				if err != nil {
					log.Error().Err(err).Send()
					writeErrResponse(r, w, http.StatusInternalServerError, "Something went wrong")
					return
				}

				if !user.Admin {
					writeErrResponse(r, w, http.StatusUnauthorized, fmt.Sprintf("user '%s' doesn't have an admin access", user.Name))
					return
				}
			}
			h.ServeHTTP(w, r)
		})
	}
}
