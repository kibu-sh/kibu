package health

import "context"

type Status struct {
	Code string `json:"code"`
}

type CheckRequest struct {
	Name string `json:"name"`
}

type CheckResponse struct {
	Value  string `json:"value"`
	Status Status `json:"status"`
}

//kibu:service
type Service interface {
	//kibu:service:method method=GET
	Check(ctx context.Context, req *CheckRequest) (res *CheckResponse, err error)
}
