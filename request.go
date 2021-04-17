package request

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Pantani/errors"
	"go.elastic.co/apm/module/apmhttp"
)

// Request object
type Request struct {
	BaseURL      string
	Headers      map[string]string
	HTTPClient   *http.Client
	ErrorHandler func(res *http.Response, uri string) error
}

// SetTimeout set the timeout request.
func (r *Request) SetTimeout(seconds time.Duration) {
	r.HTTPClient.Timeout = time.Second * seconds
}

// InitClient initialize the client request.
func InitClient(baseURL string) Request {
	return Request{
		Headers:      make(map[string]string),
		HTTPClient:   DefaultClient,
		ErrorHandler: DefaultErrorHandler,
		BaseURL:      baseURL,
	}
}

// InitJSONClient initialize the client for application/json requests.
func InitJSONClient(baseURL string) Request {
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}
	return Request{
		Headers:      headers,
		HTTPClient:   DefaultClient,
		ErrorHandler: DefaultErrorHandler,
		BaseURL:      baseURL,
	}
}

// DefaultClient represents a generic and default http client.
var DefaultClient = &http.Client{
	Timeout: time.Second * 15,
}

// DefaultErrorHandler represents a generic and default error handler.
var DefaultErrorHandler = func(res *http.Response, uri string) error {
	return nil
}

// Get sends an HTTP GET request and returns an HTTP response in background
// context, following policy (such as redirects, cookies, auth) as configured
// on the client.
//
// The response is unmarshal and stored inside the result
// parameter.
//
// An error is returned if caused by client policy (such as
// CheckRedirect), or failure to speak HTTP (such as a network
// connectivity problem). A non-2xx status code doesn't cause an
// error.
//
// eg.:
// 	var block Block
//	err := c.Get(&block, "blocks/latest", url.Values{"page": {"1"}})
//
func (r *Request) Get(result interface{}, path string, query url.Values) error {
	var queryStr = ""
	if query != nil {
		queryStr = query.Encode()
	}
	uri := strings.Join([]string{r.GetBase(path), queryStr}, "?")
	return r.Execute(context.Background(), "GET", uri, nil, result)
}

// GetWithContext sends an HTTP GET request and returns an HTTP response in the passed
// context, following policy (such as redirects, cookies, auth) as configured
// on the client.
//
// The response is unmarshal and stored inside the result
// parameter.
//
// An error is returned if caused by client policy (such as
// CheckRedirect), or failure to speak HTTP (such as a network
// connectivity problem). A non-2xx status code doesn't cause an
// error.
//
// eg.:
// 	var block Block
//	err := c.GetWithContext(&block, "blocks/latest", url.Values{"page": {"1"}}, context.Background())
//
func (r *Request) GetWithContext(ctx context.Context, result interface{}, path string, query url.Values) error {
	var queryStr = ""
	if query != nil {
		queryStr = query.Encode()
	}
	uri := strings.Join([]string{r.GetBase(path), queryStr}, "?")
	return r.Execute(ctx, "GET", uri, nil, result)
}

// Post sends an HTTP POST request and returns an HTTP response in background
// context, following policy (such as redirects, cookies, auth) as configured
// on the client.
//
// The response is unmarshal and stored inside the result
// parameter.
//
// An error is returned if caused by client policy (such as
// CheckRedirect), or failure to speak HTTP (such as a network
// connectivity problem). A non-2xx status code doesn't cause an
// error.
//
// eg.:
// 	var block Block
//	err := c.Post(&block, "blocks/latest", CustomObject{Id: 3, Name: "request"})
//
func (r *Request) Post(result interface{}, path string, body interface{}) error {
	buf, err := GetBody(body)
	if err != nil {
		return err
	}
	uri := r.GetBase(path)
	return r.Execute(context.Background(), "POST", uri, buf, result)
}

// PostWithContext sends an HTTP POST request and returns an HTTP response in the passed
// context, following policy (such as redirects, cookies, auth) as configured
// on the client.
//
// The response is unmarshal and stored inside the result
// parameter.
//
// An error is returned if caused by client policy (such as
// CheckRedirect), or failure to speak HTTP (such as a network
// connectivity problem). A non-2xx status code doesn't cause an
// error.
//
// eg.:
// 	var block Block
//	err := c.PostWithContext(&block, "blocks/latest", CustomObject{Id: 3, Name: "request"}, context.Background())
//
func (r *Request) PostWithContext(ctx context.Context, result interface{}, path string, body interface{}) error {
	buf, err := GetBody(body)
	if err != nil {
		return err
	}
	uri := r.GetBase(path)
	return r.Execute(ctx, "POST", uri, buf, result)
}

// Execute sends any HTTP request and returns an HTTP response, following
// policy (such as redirects, cookies, auth) as configured on the
// client. The response is unmarshal and stored inside the result
// parameter.
//
// An error is returned if caused by client policy (such as
// CheckRedirect), or failure to speak HTTP (such as a network
// connectivity problem). A non-2xx status code doesn't cause an
// error.
//
// eg.:
// 	var block Block
//	err := r.Execute("POST", uri, buf, result)
//
func (r *Request) Execute(ctx context.Context, method string, url string, body io.Reader, result interface{}) error {
	errParams := errors.Params{"method": method, "url": url}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return errors.E(err, errParams)
	}

	for key, value := range r.Headers {
		req.Header.Set(key, value)
	}
	c := apmhttp.WrapClient(r.HTTPClient)

	res, err := c.Do(req.WithContext(ctx))
	if err != nil {
		return errors.E(err, errParams)
	}

	err = r.ErrorHandler(res, url)
	if err != nil {
		return errors.E(err, errParams)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.E(err, errParams)
	}
	if b == nil || len(b) == 0 {
		return nil
	}
	err = json.Unmarshal(b, result)
	if err != nil {
		return errors.E(err, errParams)
	}
	return err
}

// GetBase returns the base url with path.
func (r *Request) GetBase(path string) string {
	if path == "" {
		return r.BaseURL
	}
	return fmt.Sprintf("%s/%s", r.BaseURL, path)
}

// GetBody cast custom object body to io.ReadWriter buffer
// It returns an error if occurs.
func GetBody(body interface{}) (buf io.ReadWriter, err error) {
	if body != nil {
		buf = new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(body)
	}
	return
}
