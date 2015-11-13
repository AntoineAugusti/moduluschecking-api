package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/codegangsta/negroni"
	"github.com/stretchr/testify/assert"
	"gopkg.in/redis.v3"
)

var (
	redisClient *redis.Client
)

func onStart() {
	redisClient = newRedis()
}

func onFinish() {
	redisClient.FlushAll()
	redisClient.Close()
}

func TestMiddlewareBlocksAfter5RequestsPerSecond(t *testing.T) {
	onStart()
	defer onFinish()

	n, responseRecorder := prepareLimiterMiddlewareAndRecorder()

	request, _ := http.NewRequest("GET", "foo", nil)
	request.Header.Set("Api-Key", "bar")
	n.ServeHTTP(responseRecorder, request)

	for i := 4; i >= 1; i-- {
		assert.Equal(t, strconv.Itoa(i), responseRecorder.Header().Get("Api-Remaining"))
		n.ServeHTTP(responseRecorder, request)
	}

	// See that we are blocked
	for i := 0; i < 3; i++ {
		n.ServeHTTP(responseRecorder, request)
		assertResponseWithStatusAndMessage(t, responseRecorder, 429, "rate_exceeded", "API rate exceeded. Too many requests.")
	}
}

func TestHandleAClosedRedisConnexion(t *testing.T) {
	onFinish()

	n, responseRecorder := prepareLimiterMiddlewareAndRecorder()

	request, _ := http.NewRequest("GET", "foo", nil)
	request.Header.Set("Api-Key", "bar")
	n.ServeHTTP(responseRecorder, request)

	assertResponseWithStatusAndMessage(t, responseRecorder, http.StatusInternalServerError, "server_error", "Trouble contacting Redis. Aborting.")
}

// Create a new instance of the rate limiter middleware
func newRateLimiter() *Limiter {
	return NewLimiter(redisClient)
}

// Open a new Redis connexion locally
func newRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

func prepareLimiterMiddlewareAndRecorder() (*negroni.Negroni, *httptest.ResponseRecorder) {
	n := negroni.New()
	// To be sure that we give an Api-Key to have
	// a unique identifier when accessing ressources
	n.Use(NewAuthorization())
	n.Use(newRateLimiter())

	recorder := httptest.NewRecorder()

	return n, recorder
}
