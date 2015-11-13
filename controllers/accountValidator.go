package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/AntoineAugusti/moduluschecking-api/middlewares"
	"github.com/AntoineAugusti/moduluschecking-api/responses"
	"github.com/AntoineAugusti/moduluschecking/models"
	"github.com/AntoineAugusti/moduluschecking/resolvers"
	"github.com/cloudflare/service"
	"github.com/cloudflare/service/render"
	"github.com/codegangsta/negroni"
)

type bankAccountRequest struct {
	SortCode      string `json:"sort_code"`
	AccountNumber string `json:"account_number"`
}

// Check if the bank account has got at least expected
// lengths for the sort code and the account number
func (b bankAccountRequest) isValid() bool {
	sortCodeLength := len(b.SortCode)
	accountNumberLength := len(b.AccountNumber)

	return sortCodeLength == 6 && (accountNumberLength >= 6 && accountNumberLength <= 10)
}

// Tell in JSON if a bank account is valid
type ValidityResponse struct {
	// The bank account sort code
	SortCode string `json:"sort_code"`
	// The bank account number
	AccountNumber string `json:"account_number"`
	// The validity of the given bank account
	Valid bool `json:"is_valid"`
}

type accountValidator struct {
	resolver resolvers.Resolver
}

// Handles POST for /verify
func (validator accountValidator) AccountValidatorPost(w http.ResponseWriter, req *http.Request) {
	var transformedRequest bankAccountRequest

	// Decode the JSON payload
	if err := json.NewDecoder(req.Body).Decode(&transformedRequest); err != nil {
		responses.WriteUnprocessableEntity(w)
		return
	}

	// Check that we got an expected format
	if !transformedRequest.isValid() {
		message := "Expected a 6 digits sort code and an account number between 6 and 10 digits."
		responses.WriteMessage(http.StatusBadRequest, "invalid_bank_account", message, w)
		return
	}

	// Create the required bank account
	b := models.CreateBankAccount(transformedRequest.SortCode, transformedRequest.AccountNumber)

	// Check if the bank account is valid
	isValid := validator.resolver.IsValid(b)

	// Construct and render the final response
	response := ValidityResponse{
		Valid:         isValid,
		SortCode:      b.SortCode,
		AccountNumber: b.AccountNumber,
	}
	render.JSON(w, http.StatusOK, response)
}

// Create a new account validator controller
func AccountValidatorController(parser models.Parser, limiter *middlewares.Limiter) service.WebController {
	wc := service.NewWebController("/verify")

	validator := accountValidator{
		resolver: resolvers.NewResolver(parser),
	}

	// Create middlewares to use for this route
	n := negroni.New()
	n.Use(middlewares.NewAuthorization())
	n.Use(limiter)
	n.UseHandlerFunc(validator.AccountValidatorPost)

	// Map the middlewares and the handler function to the endpoint
	wc.AddMethodHandler(service.Post, n.ServeHTTP)

	return wc
}
