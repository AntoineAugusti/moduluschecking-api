package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AntoineAugusti/moduluschecking-api/middlewares"
	"github.com/AntoineAugusti/moduluschecking-api/responses"
	"github.com/AntoineAugusti/moduluschecking/parsers"
	"github.com/cloudflare/service"
	"github.com/stretchr/testify/assert"
	"gopkg.in/redis.v3"
)

var (
	server      *httptest.Server
	base        string
	redisClient *redis.Client
)

func onStart() {
	webService := service.NewWebService()

	accountValidator := AccountValidatorController(parsers.CreateFileParser(), newRateLimiter())
	webService.AddWebController(accountValidator)

	server = httptest.NewServer(webService.BuildRouter())
	base = fmt.Sprintf("%s/verify", server.URL)
}

func onFinish() {
	redisClient.FlushAll()
}

func TestRequiresApiKey(t *testing.T) {
	defer onFinish()
	onStart()

	request, _ := http.NewRequest("POST", base, nil)
	res, _ := http.DefaultClient.Do(request)

	assertResponseWithStatusAndMessage(t, res, http.StatusUnauthorized, "authorization_required", "Please provide a HTTP header called Api-Key.")
}

func TestWarnsIfCannotDecodeJSON(t *testing.T) {
	defer onFinish()
	onStart()

	request, _ := http.NewRequest("POST", base, nil)
	request.Header.Set("Api-Key", "foo")
	res, _ := http.DefaultClient.Do(request)

	assert422Response(t, res)
}

func TestWarnsIfBankAccountDetailsAreNotValid(t *testing.T) {
	defer onFinish()
	onStart()

	res := doRequest("foo", "123456", "11225")

	message := "Expected a 6 digits sort code and an account number between 6 and 10 digits."
	assertResponseWithStatusAndMessage(t, res, http.StatusBadRequest, "invalid_bank_account", message)
}

func TestCanCheckIfABankAccountIsActuallyValid(t *testing.T) {
	defer onFinish()
	onStart()

	// A valid bank account
	res := doRequest("foo", "308037", "49743860")
	assertBankAccountResponse(t, res, "308037", "49743860", true)

	// A non valid bank account
	res = doRequest("foo", "308037", "49743861")
	assertBankAccountResponse(t, res, "308037", "49743861", false)
}

func TestRateLimitIsInPlace(t *testing.T) {
	defer onFinish()
	onStart()

	// Rate is limited at 5 request per second
	for i := 0; i < 5; i++ {
		doRequest("foo", "308037", "49743861")
	}

	res := doRequest("foo", "308037", "49743861")

	assertResponseWithStatusAndMessage(t, res, 429, "rate_exceeded", "API rate exceeded. Too many requests.")
}

func doRequest(apiKey, sortCode, accountNumber string) *http.Response {
	// Craft the JSON payload
	payload := fmt.Sprintf(`{
      "sort_code":"%s",
      "account_number": "%s"
    }`, sortCode, accountNumber)
	reader := strings.NewReader(payload)

	// Craft the request
	request, err := http.NewRequest("POST", base, reader)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Api-Key", apiKey)

	// Send the request
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}

	return response
}

// Create a new instance of the rate limiter middleware
func newRateLimiter() *middlewares.Limiter {
	redisClient = newRedis()
	return middlewares.NewLimiter(redisClient)
}

// Open a new Redis connexion locally
func newRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

func assert422Response(t *testing.T, res *http.Response) {
	assertResponseWithStatusAndMessage(t, res, 422, "invalid_json", "Cannot decode the given JSON payload.")
}

func assertBankAccountResponse(t *testing.T, res *http.Response, sortCode, accountNumber string, isValid bool) {
	var response ValidityResponse
	json.NewDecoder(res.Body).Decode(&response)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json; charset=UTF-8", res.Header.Get("Content-Type"))

	assert.Equal(t, sortCode, response.SortCode)
	assert.Equal(t, accountNumber, response.AccountNumber)
	assert.Equal(t, isValid, response.Valid)
}

func assertResponseWithStatusAndMessage(t *testing.T, res *http.Response, code int, status, message string) {
	var apiMessage responses.APIMessage
	json.NewDecoder(res.Body).Decode(&apiMessage)

	assert.Equal(t, code, res.StatusCode)
	assert.Equal(t, "application/json; charset=UTF-8", res.Header.Get("Content-Type"))

	assert.Equal(t, status, apiMessage.Status)
	assert.Equal(t, message, apiMessage.Message)
}
