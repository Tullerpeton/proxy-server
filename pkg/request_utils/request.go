package request_utils

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/proxy-server/internal/pkg/models"
)

func ParseRequest(r *http.Request, scheme string) (*models.Request, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	queryParams := make(map[string][]string)
	for key, param := range r.URL.Query() {
		queryParams[key] = param
	}
	request := &models.Request{
		Method:  r.Method,
		Scheme:  scheme,
		Host:    r.Host,
		Path:    r.URL.Path,
		Headers: r.Header,
		Params:  queryParams,

		Body: string(bodyBytes),
	}

	return request, nil
}

func ParseResponse(rp *http.Response, requestId int64) (*models.Response, error) {
	var body io.ReadCloser

	switch rp.Header.Get("Content-Encoding") {
	case "gzip":
		var err error
		body, err = gzip.NewReader(rp.Body)
		if err != nil {
			body = rp.Body
		}
	default:
		body = rp.Body
	}
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	rp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	response := &models.Response{
		RequestId: requestId,
		Code:      rp.StatusCode,
		Message:   rp.Status,
		Headers:   rp.Header,
		Body:      string(bodyBytes),
	}

	return response, nil
}
