package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AntoineAugusti/moduluschecking-api/responses"
	"github.com/codegangsta/negroni"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareBlocksWithoutApiKey(t *testing.T) {
	n, responseRecorder := prepareMiddlewareAndRecorder()

	request, _ := http.NewRequest("GET", "foo", nil)
	n.ServeHTTP(responseRecorder, request)

	assertAuthorizationRequired(t, responseRecorder)
}

func TestMiddlewareBlockWithWrongApiKey(t *testing.T) {
	n, responseRecorder := prepareMiddlewareAndRecorder()

	request, _ := http.NewRequest("GET", "bar", nil)
	request.Header.Set("Api-Key", "ab")
	n.ServeHTTP(responseRecorder, request)

	assertAuthorizationRequired(t, responseRecorder)
}

func TestMiddlewareLetThroughWithValidApiKey(t *testing.T) {
	n, responseRecorder := prepareMiddlewareAndRecorder()

	request, _ := http.NewRequest("GET", "bar", nil)
	request.Header.Set("Api-Key", "foo")
	n.ServeHTTP(responseRecorder, request)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)
}

func prepareMiddlewareAndRecorder() (*negroni.Negroni, *httptest.ResponseRecorder) {
	n := negroni.New()
	n.Use(NewAuthorization())

	recorder := httptest.NewRecorder()

	return n, recorder
}

func assertAuthorizationRequired(t *testing.T, res *httptest.ResponseRecorder) {
	status := "authorization_required"
	message := "Please provide a HTTP header called Api-Key."
	assertResponseWithStatusAndMessage(t, res, http.StatusUnauthorized, status, message)
}

func assertResponseWithStatusAndMessage(t *testing.T, res *httptest.ResponseRecorder, code int, status, message string) {
	var apiMessage responses.APIMessage
	json.NewDecoder(res.Body).Decode(&apiMessage)

	assert.Equal(t, code, res.Code)
	assert.Equal(t, "application/json; charset=UTF-8", res.Header().Get("Content-Type"))

	assert.Equal(t, status, apiMessage.Status)
	assert.Equal(t, message, apiMessage.Message)
}
