package middlewares

import (
	"net/http"

	"github.com/AntoineAugusti/moduluschecking-api/responses"
)

type Authorization struct {
}

// This middleware check that an API key was given
func NewAuthorization() *Authorization {
	return &Authorization{}
}

// The middleware handler
func (l *Authorization) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {

	apiKey := req.Header.Get("Api-Key")

	if len(apiKey) == 0 {
		responses.WriteUnauthorized(w)
		return
	}

	// Call the next middleware handler
	next(w, req)
}
