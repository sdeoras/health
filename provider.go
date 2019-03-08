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
	manager      jwt.Manager
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
		if p.manager != nil {
			// validate input request
			err := p.manager.Validate(r)
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
