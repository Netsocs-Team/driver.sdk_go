package httpx

import "github.com/go-resty/resty/v2"

// Resty returns a new resty client wired to the SDK transport, so it honours
// the TLS configuration described in the package documentation.
//
// Use it instead of resty.New() anywhere the SDK talks to the DriverHub.
func Resty() *resty.Client {
	return resty.New().SetTransport(Transport())
}
