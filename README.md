# health check for micro-services
this repo provides a convenient way to provide health check status
for your micro-services

## server side usage
```go
import "github.com/sdeoras/health"

func main() {
	h := health.New(health.OutputProto, nil, nil)
	h.Register("myService", nil)
	f := h.NewHTTPHandler()
	_ = f // f is your http handler that you can use on the server side
}
```

## client side usage
```go
import "github.com/sdeoras/health"

func main() {
	req, err := healthProvider.NewHTTPRequest("myService", "https://"+filepath.Join(
		"url",
		health.StdRoute))
    
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
    
	status, mesg, err := healthProvider.ReadResponseAndClose(resp)
	if err != nil {
		return nil, err
	}
}
```