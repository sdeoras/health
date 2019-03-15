package health

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// provider implements Provider interface
type provider struct {
	services     map[string]func(w http.ResponseWriter, r *http.Request)
	outputFormat OutputFormat
	mu           sync.Mutex
}

// Register registers upstream services. You can pass nil handler if you wish to report
// health status without any further redirection
func (p *provider) Register(service string, handler func(w http.ResponseWriter, r *http.Request)) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.services[service] = handler
}

// NewHTTPHandler returns a http handler
func (p *provider) NewHTTPHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		request := new(healthpb.HealthCheckRequest)
		response := new(healthpb.HealthCheckResponse)

		// read output format from query and override pre-initialized value
		outputFormat := p.outputFormat
		if keys, ok := r.URL.Query()[OutputFormatKey]; ok {
			switch strings.ToLower(keys[0]) {
			case string(OutputProto):
				outputFormat = OutputProto
			case string(OutputJSON):
				outputFormat = OutputJSON
			case string(OutputMesg):
				outputFormat = OutputMesg
			default:
				http.Error(w,
					fmt.Sprintf("%s", "bad request output format"),
					http.StatusBadRequest)
				return
			}
		}

		// either read from the body or from the path
		if r.Body != nil {
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
		} else {
			// get service name from the url
			if keys, ok := r.URL.Query()[ServiceKey]; ok {
				request.Service = keys[0]
			} else {
				http.Error(w,
					fmt.Sprintf("%s", "could not get service name in request body or url query"),
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
		switch outputFormat {
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

// NewHTTPRequest prepares an http request for a service to check health status
func (p *provider) NewHTTPRequest(service, url string) (*http.Request, error) {
	request := new(healthpb.HealthCheckRequest)
	request.Service = service

	b, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	return http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
}

// ReadResponseAndClose reads http response to extract health status received from server
// and closes it
func (p *provider) ReadResponseAndClose(resp *http.Response) (bool, string, error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("%s:%s. Mesg:%s",
			"expected status 200 OK, got", resp.Status, string(b))
	}

	response := new(healthpb.HealthCheckResponse)

	switch p.outputFormat {
	case OutputProto:
		if err := proto.Unmarshal(b, response); err != nil {
			return false, "", err
		} else {
			return response.Status == healthpb.HealthCheckResponse_SERVING,
				response.Status.String(), nil
		}
	case OutputJSON:
		if err := json.Unmarshal(b, response); err != nil {
			return false, "", err
		} else {
			return response.Status == healthpb.HealthCheckResponse_SERVING,
				response.Status.String(), nil
		}
	case OutputMesg:
		return string(b) == healthpb.HealthCheckResponse_SERVING.String(),
			string(b), nil
	default:
		return false, "", fmt.Errorf("invalid output format, cannot unmarshal")
	}
}

func (p *provider) SetQuery(service, rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set(ServiceKey, service)
	q.Set(OutputFormatKey, p.outputFormat.String())
	u.RawQuery = q.Encode()

	return u.String(), nil
}
