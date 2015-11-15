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
func (a *Authorization) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {

	apiKey := req.Header.Get("Api-Key")

	if len(apiKey) == 0 || !a.apiKeyExists(apiKey) {
		responses.WriteUnauthorized(w)
		return
	}

	// Call the next middleware handler
	next(w, req)
}

// Check that an API key exists. This is where in real life
// a database call should be done.
func (a Authorization) apiKeyExists(apiKey string) bool {
	return apiKey == "foo"
}
