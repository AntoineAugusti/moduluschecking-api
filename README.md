[![Travis CI](https://img.shields.io/travis/AntoineAugusti/moduluschecking-api/master.svg?style=flat-square)](https://travis-ci.org/AntoineAugusti/moduluschecking-api)
[![Software License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/AntoineAugusti/moduluschecking-api/LICENSE.md)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/AntoineAugusti/moduluschecking-api)
[![Coverage Status](http://codecov.io/github/AntoineAugusti/moduluschecking-api/coverage.svg?branch=master)](http://codecov.io/github/AntoineAugusti/moduluschecking-api?branch=master)

## Modulus checking API
This package is an API to validate UK bank account numbers, supporting authentication and rate limits. This package has been presented in a [blog post](https://blog.antoine-augusti.fr/2015/11/developing-and-deploying-a-modulus-checking-api/) where you can find instructions to deploy behind Nginx.

## What is modulus checking?
Modulus checking is a procedure for validating sort code and account number combinations for UK bank accounts.

## Requirements
- A working Go installation
- A Redis server if you want to rate limit requests

## Other packages used
For additional documentation, check out these packages I'm using:
- [https://github.com/cloudflare/service](cloudflare/service) to create the web service and provide default endpoints
- [https://github.com/etcinit/speedbump](etcinit/speedbump) for the rate limiting functionality
- [https://github.com/codegangsta/negroni](codegangsta/negroni) for middlewares
- [https://github.com/AntoineAugusti/moduluschecking](AntoineAugusti/moduluschecking) for actually validating bank account numbers

## Getting started
You can grab this package with the following command:
```
go get github.com/AntoineAugusti/moduluschecking-api
```

And then build it using the Makefile:
```
cd ${GOPATH%/}/src/github.com/AntoineAugusti/moduluschecking-api
make build
cp moduluschecking-api ${GOPATH%/}/bin/moduluschecking-api
```

Building the package using the Makefile will allow the `/heartbeat` and `/version` endpoints to echo the Git hash of the current build and the date of the current build.

## Usage
From the `-h` flag:
```
Usage of ./moduluschecking-api:
  -a string
        address to listen (default ":8080")
```

## API endpoint
### `/verify`
**Purpose:** Check that a UK bank account number is valid.

#### Request
Expected HTTP headers:

Key  | Value
------------- | -------------
`Api-Key`  | The value of the API key. The only supported value for now is `foo`.
`Content-Type`  | `application/json; charset=UTF-8`

JSON payload:
```json
{
  "sort_code":"123456",
  "account_number":"12345678"
}
```

#### Response
##### Success
For a status code of **200**:

HTTP header key  | Value
------------- | -------------
`Api-Remaining`  | The number of requests remaining. The current limit is 5 requests / second. Example: `3`
`Content-Type`  | `application/json; charset=UTF-8`

Body:
```json
{
  "sort_code":"308037",
  "account_number":"87344782",
  "is_valid":false
}
```

##### Errors
Each error is given as a JSON response containing 2 keys (`status` and `message`), describing the source of the error. The `Content-Type` HTTP header is always set to `application/json; charset=UTF-8`.

Example:
```json
{
  "status":"rate_exceeded",
  "message":"API rate exceeded. Too many requests."
}
```

HTTP status code  | `status` value | `message` value
:-------------: | ------------- | -------------
400 | `invalid_bank_account` | `Expected a 6 digits sort code and an account number between 6 and 10 digits.`
401 | `authorization_required` | `Please provide a HTTP header called Api-Key`
422 | `invalid_json` | `Cannot decode the given JSON payload`
429 | `rate_exceeded` | `API rate exceeded. Too many requests.`
