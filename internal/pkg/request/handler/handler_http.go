package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/proxy-server/internal/pkg/request"
	"github.com/proxy-server/pkg/proxy"
	"github.com/proxy-server/pkg/scaner"

	"github.com/gorilla/mux"
)

type RequestHandler struct {
	requestUseCase request.UseCase
	proxyManager   *proxy.ProxyManager
}

func NewHandler(requestUseCase request.UseCase, proxyManager *proxy.ProxyManager) request.Handler {
	return &RequestHandler{
		requestUseCase: requestUseCase,
		proxyManager:   proxyManager,
	}
}

func (h *RequestHandler) ProxyRequest(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Method == http.MethodConnect {
		err = h.proxyManager.ProxyHttpsRequest(w, r)
	} else {
		err = h.proxyManager.ProxyHttpRequest(w, r)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *RequestHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 32)
	selectedRequest, err := h.requestUseCase.GetRequestDataById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	answer, err := json.Marshal(selectedRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(answer)
}

func (h *RequestHandler) GetRequests(w http.ResponseWriter, r *http.Request) {
	selectedRequests, err := h.requestUseCase.GetAllRequestsData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	answer, err := json.Marshal(selectedRequests)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(answer)
}

func (h *RequestHandler) RepeatRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 32)
	selectedRequest, err := h.requestUseCase.GetRequestById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	h.ProxyRequest(w, &http.Request{
		Method: selectedRequest.Method,
		URL: &url.URL{
			Scheme: selectedRequest.Scheme,
			Host:   selectedRequest.Host,
			Path:   selectedRequest.Path,
		},
		Header: selectedRequest.Headers,
		Body:   ioutil.NopCloser(strings.NewReader(selectedRequest.Body)),
		Host:   r.Host,
	})
}

func (h *RequestHandler) ScanRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 32)
	selectedRequest, err := h.requestUseCase.GetRequestById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	request := &http.Request{
		Method: selectedRequest.Method,
		URL: &url.URL{
			Scheme: selectedRequest.Scheme,
			Host:   selectedRequest.Host,
			Path:   selectedRequest.Path,
		},
		Header: selectedRequest.Headers,
		Body:   ioutil.NopCloser(strings.NewReader(selectedRequest.Body)),
		Host:   r.Host,
	}
	params := url.Values{}
	for key, values := range selectedRequest.Params {
		for _, val := range values {
			params.Add(key, val)
		}
	}

	var flag bool
	for _, val := range scaner.GetParams() {
		randValue := scaner.RandStringRunes()
		newParams := params
		newParams.Add(val, randValue)
		request.URL.RawQuery = newParams.Encode()
		resp, err := http.DefaultTransport.RoundTrip(request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if strings.Contains(string(body), randValue) {
			w.Write([]byte(val + "-найден скрытый гет параметр\n"))
			flag = true
		}
	}
	if flag == false {
		w.Write([]byte("скрытые гет параметры не найдены\n"))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
}
