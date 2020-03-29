package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Pantani/errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Request the request object
type Request struct {
	BaseUrl      string
	Headers      map[string]string
	HttpClient   *http.Client
	ErrorHandler func(res *http.Response, uri string) error
}

// SetTimeout set the default request timeout
func (r *Request) SetTimeout(seconds time.Duration) {
	r.HttpClient.Timeout = time.Second * seconds
}

// InitClient initialize the client
// It returns the request object
func InitClient(baseUrl string) Request {
	return Request{
		Headers:      make(map[string]string),
		HttpClient:   DefaultClient,
		ErrorHandler: DefaultErrorHandler,
		BaseUrl:      baseUrl,
	}
}

var DefaultClient = &http.Client{
	Timeout: time.Second * 15,
}

var DefaultErrorHandler = func(res *http.Response, uri string) error {
	return nil
}

// Get make a GET request with query url values.
// It returns an error if occurs.
func (r *Request) Get(result interface{}, path string, query url.Values) error {
	var queryStr = ""
	if query != nil {
		queryStr = query.Encode()
	}
	uri := strings.Join([]string{r.getUrl(path), queryStr}, "?")
	return r.Execute("GET", uri, nil, result, errors.Params{"path": path})
}

// Get make a POST request with body value.
// It returns an error if occurs.
func (r *Request) Post(result interface{}, path string, body interface{}) error {
	buf, err := getBody(body)
	if err != nil {
		return err
	}
	uri := r.getUrl(path)
	return r.Execute("POST", uri, buf, result, errors.Params{"path": path})
}

// Get make a custom request with body and query url values.
// It returns an error if occurs.
func (r *Request) Execute(method string, url string, body io.Reader, result interface{}, params errors.Params) error {
	params["method"] = method
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return errors.E(err, params)
	}

	for key, value := range r.Headers {
		req.Header.Set(key, value)
	}
	res, err := r.HttpClient.Do(req)
	if err != nil {
		return errors.E(err, params)
	}

	err = r.ErrorHandler(res, url)
	if err != nil {
		return errors.E(err, params)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.E(err, params)
	}
	err = json.Unmarshal(b, result)
	if err != nil {
		return errors.E(err, params)
	}
	return err
}

// getUrl join the path with url.
// It returns the full url.
func (r *Request) getUrl(path string) string {
	if path == "" {
		return fmt.Sprintf("%s", r.BaseUrl)
	}
	return fmt.Sprintf("%s/%s", r.BaseUrl, path)
}

// getBody convert a generic interface into a ReadWriter object.
// It returns the buffer and an error if occurs.
func getBody(body interface{}) (buf io.ReadWriter, err error) {
	if body != nil {
		buf = new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(body)
	}
	return
}
