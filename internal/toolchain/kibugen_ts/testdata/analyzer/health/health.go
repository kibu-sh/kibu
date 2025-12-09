package health

import "context"

type Status struct {
	Code string `json:"code"`
}

type CheckRequest struct {
	Name string `json:"name"`
}

type CheckResponse struct {
	Value          string   `json:"value"`
	Status         Status   `json:"status"`
	StatusList     []Status `json:"status_list"`
	OptionalStatus *Status  `json:"optional_status"`
}

//kibu:service
type ServiceV1 interface {
	//kibu:service:method method=GET
	Check(ctx context.Context, req *CheckRequest) (res *CheckResponse, err error)
}

//kibu:service
type ServiceV2 interface {
	//kibu:service:method method=GET
	Check(ctx context.Context, req *CheckRequest) (res *CheckResponse, err error)
}
