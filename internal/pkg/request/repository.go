package request

import "github.com/proxy-server/internal/pkg/models"

type Repository interface {
	GetAllRequestsData() ([]*models.RequestData, error)
	GetRequestById(id int64) (*models.Request, error)
	GetRequestDataById(id int64) (*models.RequestData, error)
	InsertRequest(request *models.Request) error
	InsertResponse(response *models.Response) error
}
