package responses

import (
	"net/http"

	"github.com/cloudflare/service/render"
)

// A simple struct to easily respond some JSON
type APIMessage struct {
	// The HTTP status code
	code int
	// The status key
	Status string `json:"status"`
	// A human-readable message
	Message string `json:"message"`
}

// Write a message saying that the client should provide
// authentication details
func WriteUnauthorized(w http.ResponseWriter) {
	WriteMessage(401, "authorization_required", "Please provide a HTTP header called Api-Key.", w)
}

// Write a message saying that we cannot parse the given JSON payload
func WriteUnprocessableEntity(w http.ResponseWriter) {
	WriteMessage(422, "invalid_json", "Cannot decode the given JSON payload.", w)
}

// Write a message with a given status code, a status and a message
func WriteMessage(code int, status string, message string, w http.ResponseWriter) {
	response := APIMessage{
		code:    code,
		Status:  status,
		Message: message,
	}
	render.JSON(w, response.code, response)
}
