package health

import (
	"net/http"

	"github.com/sdeoras/jwt"
)

type OutputFormat int

const (
	OutputProto OutputFormat = iota
	OutputJSON
	OutputMesg
)

const (
	StdRoute = "/health"
)

// Provider makes an http func for health check
type Provider interface {
	// Register registers upstream services and corresponding redirection.
	// Pass nil handler if you wish to provide health status without further redirection.
	Register(service string, handler func(w http.ResponseWriter, r *http.Request))

	// server side ---------
	// Provide provides an http handler
	Provide() func(w http.ResponseWriter, r *http.Request)

	// client side ---------
	// Request prepares an http request for a service to check health status
	Request(service, url string) (*http.Request, error)
	// Response gets a health response status, string and error from http response
	Response(resp *http.Response) (bool, string, error)
}

// New provides an instance of health check func Provider.
// It takes jwt validator if jwt auth is required. Pass nil for jwt operator if not required.
func NewProvider(outputFormat OutputFormat,
	jwtManager jwt.Manager,
	jwtClaims map[string]interface{}) Provider {
	p := new(provider)
	p.outputFormat = outputFormat
	p.jwtManager = jwtManager
	p.jwtClaims = jwtClaims
	p.services = make(map[string]func(w http.ResponseWriter, r *http.Request))
	return p
}
