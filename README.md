[![Go Reference](https://pkg.go.dev/badge/github.com/Pantani/request.svg)](https://pkg.go.dev/github.com/Pantani/request)
[![codecov](https://codecov.io/gh/Pantani/request/branch/master/graph/badge.svg?token=BNDBT0HFFT)](https://codecov.io/gh/Pantani/request)

# Client request abstraction

Simple abstraction for client requests with memory cache.

Initialize the client:
```go
import "github.com/Pantani/request"

client := request.InitClient("http://127.0.0.1:8080")
// OR
client := request.Request{
	HttpClient:   request.DefaultClient,
	ErrorHandler: request.DefaultErrorHandler,
	BaseUrl:      "http://127.0.0.1:8080",
	Headers: map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	},
}
```
## Methods

### GET

```go
var result CustomResult
err := client.Get(&result, "api/v1/object", url.Values{"id": {"69"}})

// with cache
err := request.GetWithCache(&result, "api/v1/object", url.Values{"id": {"69"}}, time.Hour*1)
```

### POST

```go
var result CustomResult
err := client.Post(&result, "api/v1/object", Request{Name: "name", Id: "id"})

// with cache
err := request.PostWithCache(&result, "api/v1/object", Request{Name: "name", Id: "id"}, time.Hour*1)
```

## Parameters

- Add Error Handler:
```go
client.ErrorHandler = func(res *http.Response, desc string) error {
	switch res.StatusCode {
	case http.StatusBadRequest:
		return getAPIError(res, desc)
	case http.StatusNotFound:
		return blockatlas.ErrNotFound
	case http.StatusOK:
		return nil
	default:
		return errors.E("getHTTPError error", errors.Params{"status": res.Status})
	}
}
```

- Set timeout:
```go
client.SetTimeout(35)
```

- Add header:
```go
client.Headers["X-API-KEY"] = "<API_KEY>"
```
