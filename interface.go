package health

import (
	"net/http"
)

// Provider provides ways to send and receive health status for your http micro-services
type Provider interface {
	// server side ---------

	// Register registers upstream services and corresponding redirection.
	// Pass nil handler if you wish to provide health status without further redirection.
	Register(service string, handler func(w http.ResponseWriter, r *http.Request))
	// NewHTTPHandler provides a new http handler
	NewHTTPHandler() func(w http.ResponseWriter, r *http.Request)

	// client side ---------

	// NewHTTPRequest prepares an http request for a service to check health status
	NewHTTPRequest(service, url string) (*http.Request, error)
	// SetQuery sets query in input url producing modified url
	SetQuery(service, rawurl string) (string, error)
	// ReadResponseAndClose gets a health response status, string and error from http response
	ReadResponseAndClose(resp *http.Response) (bool, string, error)
}

// New provides an instance of health check func Provider.
func NewProvider(outputFormat OutputFormat) Provider {
	p := new(provider)
	p.outputFormat = outputFormat
	p.services = make(map[string]func(w http.ResponseWriter, r *http.Request))
	return p
}
