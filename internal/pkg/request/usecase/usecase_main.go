package usecase

import (
	"github.com/proxy-server/internal/pkg/models"
	"github.com/proxy-server/internal/pkg/request"
)

type ProxyUseCase struct {
	proxyRepository request.Repository
}

func NewProxyUseCase(proxyRepository request.Repository) request.UseCase {
	return &ProxyUseCase{
		proxyRepository: proxyRepository,
	}
}

func (u *ProxyUseCase) GetRequestDataById(id int64) (*models.RequestData, error) {
	return u.proxyRepository.GetRequestDataById(id)
}

func (u *ProxyUseCase) GetRequestById(id int64) (*models.Request, error) {
	return u.proxyRepository.GetRequestById(id)
}

func (u *ProxyUseCase) GetAllRequestsData() ([]*models.RequestData, error) {
	return u.proxyRepository.GetAllRequestsData()
}

func (u *ProxyUseCase) SaveRequest(request *models.Request) error {
	return u.proxyRepository.InsertRequest(request)
}

func (u *ProxyUseCase) ScanRequest() {
	panic("implement me")
}
