package health

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/proto"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// TestMaker_Register_OutputProto checks if proto output from a registered
// health check service comes out as expected
func TestMaker_Register_OutputProto(t *testing.T) {
	service := "my-service"

	m := NewProvider(OutputProto, nil, nil)
	m.Register(service, nil)
	f := m.Provide()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(f)

	request := &healthpb.HealthCheckRequest{
		Service: service,
	}
	response := new(healthpb.HealthCheckResponse)

	b, err := proto.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "/", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	b, err = ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}

	if err := proto.Unmarshal(b, response); err != nil {
		t.Fatal(err)
	}

	if response.Status != healthpb.HealthCheckResponse_SERVING {
		t.Fatal("expected", healthpb.HealthCheckResponse_SERVING, "got", response.Status)
	}
}

// TestMaker_Register_OutputJSON checks if json output from a registered
// health check service comes out as expected
func TestMaker_Register_OutputJSON(t *testing.T) {
	service := "my-service"

	m := NewProvider(OutputJSON, nil, nil)
	m.Register(service, nil)
	f := m.Provide()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(f)

	request := &healthpb.HealthCheckRequest{
		Service: service,
	}
	response := new(healthpb.HealthCheckResponse)

	b, err := proto.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "/", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	b, err = ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(b, response); err != nil {
		t.Fatal(err)
	}

	if response.Status != healthpb.HealthCheckResponse_SERVING {
		t.Fatal("expected", healthpb.HealthCheckResponse_SERVING, "got", response.Status)
	}
}

// TestMaker_Register_OutputMesg checks if mesg output from a registered
// health check service comes out as expected
func TestMaker_Register_OutputMesg(t *testing.T) {
	service := "my-service"

	m := NewProvider(OutputMesg, nil, nil)
	m.Register(service, nil)
	f := m.Provide()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(f)

	request := &healthpb.HealthCheckRequest{
		Service: service,
	}

	b, err := proto.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "/", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	b, err = ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != healthpb.HealthCheckResponse_SERVING.String() {
		t.Fatal("expected", healthpb.HealthCheckResponse_SERVING, "got", string(b))
	}
}

// TestMaker_Register_Redirect_OutputProto checks if upstream redirection works
// as expected
func TestMaker_Register_Redirect_OutputProto(t *testing.T) {
	service := "my-service"

	mRedirect := NewProvider(OutputProto, nil, nil)
	mRedirect.Register(service, nil)

	m := NewProvider(OutputMesg, nil, nil)
	m.Register(service, mRedirect.Provide())
	f := m.Provide()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(f)

	request := &healthpb.HealthCheckRequest{
		Service: service,
	}
	response := new(healthpb.HealthCheckResponse)

	b, err := proto.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "/", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	b, err = ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}

	if err := proto.Unmarshal(b, response); err != nil {
		t.Fatal(err)
	}

	if response.Status != healthpb.HealthCheckResponse_SERVING {
		t.Fatal("expected", healthpb.HealthCheckResponse_SERVING, "got", response.Status)
	}
}

// TestMaker_NotRegistered_OutputProto checks if output from a non-registered
// service comes out as expected
func TestMaker_NotRegistered_OutputProto(t *testing.T) {
	service := "my-service"

	m := NewProvider(OutputProto, nil, nil)
	f := m.Provide()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(f)

	request := &healthpb.HealthCheckRequest{
		Service: service,
	}
	response := new(healthpb.HealthCheckResponse)

	b, err := proto.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "/", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	b, err = ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}

	if err := proto.Unmarshal(b, response); err != nil {
		t.Fatal(err)
	}

	if response.Status != healthpb.HealthCheckResponse_SERVICE_UNKNOWN {
		t.Fatal("expected", healthpb.HealthCheckResponse_SERVICE_UNKNOWN, "got", response.Status)
	}
}
