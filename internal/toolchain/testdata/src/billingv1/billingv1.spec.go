package billingv1

import (
	"context"
	"go.temporal.io/sdk/workflow"
)

//type AccountStatus string
//
//const (
//	AccountStatusUnsubscribed   AccountStatus = "trial"
//	AccountStatusSubscribed     AccountStatus = "subscribed"
//	AccountStatusPaymentFailed  AccountStatus = "payment_failed"
//	AccountStatusPaymentPending AccountStatus = "payment_pending"
//	AccountStatusCanceled       AccountStatus = "canceled"
//)

type WatchAccountRequest struct{}

type WatchAccountResponse struct {
	//Status AccountStatus
}

type ChargePaymentMethodRequest struct {
	Fail bool `json:"fail"`
}

type ChargePaymentMethodResponse struct {
	Success bool `json:"success"`
}

type CustomerSubscriptionsRequest struct{}
type CustomerSubscriptionsResponse struct{}

type SetDiscountRequest struct {
	DiscountCode string `json:"discount_code"`
}

type CancelBillingRequest struct{}

type AttemptPaymentRequest struct {
	Fail bool `json:"fail"`
}

type AttemptPaymentResponse struct {
	Success bool `json:"success"`
}

type GetAccountDetailsRequest struct{}
type GetAccountDetailsResponse struct {
	//Status AccountStatus
}

// Service is the public-facing API for this system
//
//kibu:service public
type Service interface {
	// WatchAccount watches the account status
	//
	//kibu:service:method
	WatchAccount(ctx context.Context, req WatchAccountRequest) (res WatchAccountResponse, err error)
}

// Activities synchronize the workflow state with an external payment gateway
//
//kibu:activity task_queue=payments
type Activities interface {
	// ChargePaymentMethod performs work against another transactional system
	//
	//kibu:activity:method
	ChargePaymentMethod(ctx context.Context, req ChargePaymentMethodRequest) (res ChargePaymentMethodResponse, err error)
}

// CustomerSubscriptionsWorkflow represents a single long-running workflow for a customer
//
//kibu:workflow task_queue=payments
type CustomerSubscriptionsWorkflow interface {
	// Execute initiates a long-running workflow for the customers account
	//
	//kibu:workflow:execute
	Execute(ctx workflow.Context, req CustomerSubscriptionsRequest) (res CustomerSubscriptionsResponse, err error)

	// AttemptPayment attempts to charge the customers payment method
	// the account status will reflect the outcome of the attempt
	//
	//kibu:workflow:update
	AttemptPayment(ctx workflow.Context, req AttemptPaymentRequest) (res AttemptPaymentResponse, err error)

	// SetDiscount sets the discount code for the customer
	//
	//kibu:workflow:signal
	SetDiscount(ctx workflow.Context, req SetDiscountRequest) error

	// CancelBilling sends a signalChannel to cancel the customer's billing process
	// this will end the workflow
	//
	//kibu:workflow:signal
	CancelBilling(ctx workflow.Context, req CancelBillingRequest) error

	// GetAccountDetails returns the current account status
	// should not mutate state, doesn't have context
	// should not call activities (helps enforce determinism)
	//
	//kibu:workflow:query
	GetAccountDetails(req GetAccountDetailsRequest) (res GetAccountDetailsResponse, err error)
}
