package main

import (
	"flag"

	"github.com/AntoineAugusti/moduluschecking-api/controllers"
	"github.com/AntoineAugusti/moduluschecking-api/middlewares"
	"github.com/AntoineAugusti/moduluschecking/parsers"
	"github.com/cloudflare/service"
	"gopkg.in/redis.v3"
)

var buildTag = "dev"
var buildDate = "0001-01-01T00:00:00Z"

func main() {
	service.BuildTag = buildTag
	service.BuildDate = buildDate
	service.VersionRoute = "/version"
	service.HeartbeatRoute = "/heartbeat"

	address := flag.String("a", ":8080", "address to listen")
	flag.Parse()

	webService := service.NewWebService()

	parser := parsers.CreateFileParser()
	accountValidator := controllers.AccountValidatorController(parser, newRateLimiter())
	webService.AddWebController(accountValidator)

	webService.Run(*address)
}

// Create a new instance of the rate limiter middleware
func newRateLimiter() *middlewares.Limiter {
	return middlewares.NewLimiter(newRedis())
}

// Open a new Redis connexion locally
func newRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}
