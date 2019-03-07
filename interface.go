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
	// Provide provides an http handler
	Provide() func(w http.ResponseWriter, r *http.Request)
	// Register registers upstream services and corresponding redirection.
	// Pass nil handler if you wish to provide health status without further redirection.
	Register(service string, handler func(w http.ResponseWriter, r *http.Request))
}

// New provides an instance of health check func Provider.
// It takes jwt validator if jwt auth is required. Pass nil for jwt operator if not required.
func New(outputFormat OutputFormat, validator jwt.Validator) Provider {
	m := new(provider)
	m.outputFormat = outputFormat
	m.validator = validator
	m.services = make(map[string]func(w http.ResponseWriter, r *http.Request))
	return m
}
