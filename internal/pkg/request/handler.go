package request

import (
	"net/http"
)

type Handler interface {
	ProxyRequest(w http.ResponseWriter, r *http.Request)
	GetRequest(w http.ResponseWriter, r *http.Request)
	GetRequests(w http.ResponseWriter, r *http.Request)
	RepeatRequest(w http.ResponseWriter, r *http.Request)
	ScanRequest(w http.ResponseWriter, r *http.Request)
}
