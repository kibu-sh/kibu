package health

import "context"

type CheckRequest struct {
	Name string `json:"name"`
}

type CheckResponse struct {
	Status string `json:"status"`
}

//kibu:service
type Service interface {
	//kibu:service:method method=GET
	Check(ctx context.Context, req *CheckRequest) (res *CheckResponse, err error)
}
