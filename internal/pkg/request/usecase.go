package request

import "github.com/proxy-server/internal/pkg/models"

type UseCase interface {
	GetRequestById(id int64) (*models.Request, error)
	GetRequestDataById(id int64) (*models.RequestData, error)
	GetAllRequestsData() ([]*models.RequestData, error)
	SaveRequest(request *models.Request) error
	ScanRequest()
}
