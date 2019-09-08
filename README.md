# health check for micro-services
this repo provides a convenient way to provide health check status
for your serverless http micro-services

## server side usage
```go
import "github.com/sdeoras/health"

func main() {
	h := health.New(health.OutputProto)
	h.Register("myService", nil)
	f := h.NewHTTPHandler()
	_ = f // f is your http handler that you can use on the server side
	
	// typically this is served at /health route
}
```

## client side usage
### Using POST
```go
import "github.com/sdeoras/health"

func main() {
	// create a new health provider on the client side to work with proto output format.
	// i.e., server sends health status as a proto binary mesg
	h := health.New(health.OutputProto)
	
	// Obtain a new http request for a service (note that this service is something that
	// server understands since it has been registered on the server side.
	// "rawurl" is the http endpoint at which server is serving health status.
	req, err := h.NewHTTPRequest("myService", rawurl)
    
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
    
	// read response and close the response body
	status, mesg, err := h.ReadResponseAndClose(resp)
	if err != nil {
		return nil, err
	}
	
	// status is one of the variabled defined in:
	// https://godoc.org/google.golang.org/grpc/health/grpc_health_v1#pkg-variables
}
```

### Using GET
```go
import "github.com/sdeoras/health"

func main() {
	// create a new health provider on the client side to work with proto output format.
	// i.e., server sends health status as a proto binary mesg
	h := health.New(health.OutputMesg)
	
	rawurl := "http://your.domain.com/health"
	rawurl, err := h.SetQuery("myService", rawurl)
	if err != nil { /* do something */ }
	
	// rawurl now contains query parameters allowing you to request health status
	b, err := http.Get(rawurl)
	if err != nil { /* do something */ }
	
	fmt.Println(string(b))
}
```