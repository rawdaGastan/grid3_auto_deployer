// Package middlewares for middleware between api and backend
package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/rawdaGastan/cloud4students/models"
	"github.com/rs/zerolog/log"
)

// UserIDKey key saved in request context
type UserIDKey string

// Authorization to authorize users in requests
func Authorization(excludedRoutes []*mux.Route, secret string, timeout int) func(http.Handler) http.Handler {
	// Cache the regex object of each route (obviously for performance purposes)
	var excludedRoutesRegexp []*regexp.Regexp
	rl := len(excludedRoutes)
	for i := 0; i < rl; i++ {
		r := excludedRoutes[i]
		pathRegexp, _ := r.GetPathRegexp()
		regx, _ := regexp.Compile(pathRegexp)
		excludedRoutesRegexp = append(excludedRoutesRegexp, regx)
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			exclude := false
			requestMethod := r.Method
			for i := 0; i < rl; i++ {
				excludedRoute := excludedRoutes[i]
				methods, _ := excludedRoute.GetMethods()
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
					if excludedRoutesRegexp[i].MatchString(uri) {
						exclude = true
						break
					}
				}
			}
			if !exclude {
				reqToken := r.Header.Get("Authorization")
				splitToken := strings.Split(reqToken, "Bearer ")
				if len(splitToken) != 2 {
					WriteUnAuthorizedResponse(w)
					return
				}
				reqToken = splitToken[1]

				claims, err := validateToken(reqToken, secret, timeout)
				if err != nil {
					WriteUnAuthorizedResponse(w)
					return
				}
				ctx := context.WithValue(r.Context(), UserIDKey("UserID"), claims.UserID)
				h.ServeHTTP(w, r.WithContext(ctx))
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}
}

func validateToken(token, secret string, timeout int) (models.Claims, error) {
	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return models.Claims{}, err
	}
	if !tkn.Valid {
		return models.Claims{}, fmt.Errorf("token '%s' is invalid", token)
	}

	if time.Until(claims.ExpiresAt.Time) > time.Duration(timeout)*time.Minute {
		return models.Claims{}, fmt.Errorf("token '%s' is expired", token)
	}

	return *claims, nil
}

// WriteUnAuthorizedResponse write error messages in api
func WriteUnAuthorizedResponse(w http.ResponseWriter) {
	jsonErrRes, _ := json.Marshal(map[string]string{"err": "User is not authorized"})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, err := w.Write(jsonErrRes)
	if err != nil {
		log.Error().Err(err).Msg("write auth error response failed")
	}
}
