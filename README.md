# health check for micro-services
this repo provides a convenient way to provide health check status
for your micro-services

# usage
```go
import "github.com/sdeoras/health"

func main() {
	h := health.New(health.OutputProto, nil, nil)
	h.Register("myService", nil)
	f := h.Provide()
	_ = f // f is your http handler
}
```

if upstream http redirection is required for a health check, pass
a `http handler` against a service name in call to `Register`