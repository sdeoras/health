package health

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/sdeoras/jwt"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type provider struct {
	services     map[string]func(w http.ResponseWriter, r *http.Request)
	outputFormat OutputFormat
	jwtManager   jwt.Manager
	jwtClaims    map[string]interface{}
	mu           sync.Mutex
}

// Register registers upstream services. You can pass nil handler if you wish to report
// health status without any further redirection
func (p *provider) Register(service string, handler func(w http.ResponseWriter, r *http.Request)) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.services[service] = handler
}

// Provide returns a http handler
func (p *provider) Provide() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if p.jwtManager != nil {
			// validate input request
			err := p.jwtManager.Validate(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		request := new(healthpb.HealthCheckRequest)
		response := new(healthpb.HealthCheckResponse)

		// read request from body
		if b, err := ioutil.ReadAll(r.Body); err != nil {
			http.Error(w,
				fmt.Sprintf("%s:%v", "could not read from http request", err),
				http.StatusBadRequest)
			return
		} else {
			if err := proto.Unmarshal(b, request); err != nil {
				http.Error(w,
					fmt.Sprintf("%s:%v", "could not unmarshal request", err),
					http.StatusBadRequest)
				return
			}
		}

		// check if request needs to be forwarded to upstream service health check
		// handlers
		if redirect, ok := p.services[request.Service]; ok {
			if redirect != nil {
				if b, err := proto.Marshal(request); err != nil {
					http.Error(w,
						fmt.Sprintf("%s:%v", "could not marshal request", err),
						http.StatusBadRequest)
				} else {
					r.Body = ioutil.NopCloser(bytes.NewReader(b))
					redirect(w, r)
				}
				return
			} else {
				response.Status = healthpb.HealthCheckResponse_SERVING
			}
		} else {
			response.Status = healthpb.HealthCheckResponse_SERVICE_UNKNOWN
		}

		// prepare output per format
		switch p.outputFormat {
		case OutputProto:
			if b, err := proto.Marshal(response); err != nil {
				http.Error(w,
					fmt.Sprintf("%s:%v", "could not marshal response to proto", err),
					http.StatusInternalServerError)
			} else {
				_, _ = w.Write(b)
			}
		case OutputJSON:
			if b, err := json.Marshal(response); err != nil {
				http.Error(w,
					fmt.Sprintf("%s:%v", "could not marshal response to json", err),
					http.StatusInternalServerError)
			} else {
				_, _ = w.Write(b)
			}
		case OutputMesg:
			_, _ = fmt.Fprintf(w, "%s", response.Status.String())
		}
	}
}

// Request prepares an http request for a service to check health status
func (p *provider) Request(service, url string) (*http.Request, error) {
	request := new(healthpb.HealthCheckRequest)
	request.Service = service

	b, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	if p.jwtManager != nil {
		if r, err := p.jwtManager.Request(http.MethodPost, url, p.jwtClaims, b); err != nil {
			return nil, err
		} else {
			return r, nil
		}
	} else {
		r, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
		if err != nil {
			return nil, err
		}

		return r, nil
	}
}

func (p *provider) Response(resp *http.Response) (string, error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s:%s. Mesg:%s",
			"expected status 200 OK, got", resp.Status, string(b))
	}

	response := new(healthpb.HealthCheckResponse)

	switch p.outputFormat {
	case OutputProto:
		if err := proto.Unmarshal(b, response); err != nil {
			return "", err
		} else {
			return response.Status.String(), nil
		}
	case OutputJSON:
		if err := json.Unmarshal(b, response); err != nil {
			return "", err
		} else {
			return response.Status.String(), nil
		}
	case OutputMesg:
		return string(b), nil
	default:
		return "", fmt.Errorf("invalid output format, cannot unmarshal")
	}
}
