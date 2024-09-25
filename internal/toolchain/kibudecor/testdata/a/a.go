package a

import "github.com/kibu-sh/kibu/pkg/transport"

// Service implementation documents
//
//kibu:service
type Service struct{}

// ProviderFunc provider func docs
//
//kibu:provider
func ProviderFunc() Service {
	return Service{}
}

// Activities should include activity docs
//
//kibu:activity
type Activities interface {
	// DoWork
	//
	//kibu:activity:method
	DoWork() error
}

// Workflows
//
//kibu:workflow
type Workflows interface {
	// DoWork
	//
	//kibu:workflow:execute
	DoWork() error
}

// Middleware
//
//kibu:middleware
type Middleware struct{}

// GlobalMiddleware
//
//kibu:middleware tag=auth
func (m *Middleware) GlobalMiddleware(tctx transport.Context, next transport.Handler) error {
	return nil
}

// should also find something without full doc comments

//kibu:provider
func DoSomething() {}
