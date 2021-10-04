package proxy

import (
	"io"
	"net/http"

	"github.com/proxy-server/pkg/request_utils"
)

func (p *ProxyManager) ProxyHttpRequest(w http.ResponseWriter, r *http.Request) error {
	request, err := request_utils.ParseRequest(r, "http")
	if err != nil {
		return err
	}

	if err = p.proxyRepository.InsertRequest(request); err != nil {
		return err
	}

	proxyRequest, err := p.createProxyHttpRequest(r)
	if err != nil {
		return err
	}

	proxyResponse, err := p.makeHttpRequest(proxyRequest)
	if err != nil {
		return err
	}
	defer proxyResponse.Body.Close()

	if err = p.saveProxyHttpResponse(w, proxyResponse); err != nil {
		return err
	}

	response, err := request_utils.ParseResponse(proxyResponse, request.Id)
	if err != nil {
		return err
	}

	if err = p.proxyRepository.InsertResponse(response); err != nil {
		return err
	}

	return nil
}

func (p *ProxyManager) createProxyHttpRequest(r *http.Request) (*http.Request, error) {
	proxyRequest, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		return nil, err
	}
	proxyRequest.Header = r.Header

	return proxyRequest, nil
}

func (p *ProxyManager) makeHttpRequest(r *http.Request) (*http.Response, error) {
	proxyResponse, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	return proxyResponse, nil
}

func (p *ProxyManager) saveProxyHttpResponse(w http.ResponseWriter, response *http.Response) error {
	for header, values := range response.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}
	w.WriteHeader(response.StatusCode)
	_, err := io.Copy(w, response.Body)
	if err != nil {
		return err
	}

	return nil
}
