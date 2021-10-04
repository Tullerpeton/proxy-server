package proxy

import (
	"github.com/proxy-server/internal/pkg/request"
)

type ProxyManager struct {
	proxyRepository request.Repository
}

func NewProxyManager(proxyRepository request.Repository) *ProxyManager {
	return &ProxyManager{
		proxyRepository: proxyRepository,
	}
}
