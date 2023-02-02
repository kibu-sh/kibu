package templates

import (
	_ "embed"
	"github.com/discernhq/devx/internal/codedef"
)

//go:embed fixtures/worker.go.tmpl
var worker string
var Worker = MustParse[codedef.Module](DefaultOptions("worker.go.tmpl", worker))

//go:embed fixtures/service.go.tmpl
var service string
var Service = MustParse[codedef.Module](DefaultOptions("service.go.tmpl", service))

//go:embed fixtures/http_handler_factories.go.tmpl
var httpHandlerFactories string
var HttpHandlerFactoryContainer = MustParse[codedef.FactoryContainer](DefaultOptions("http_handler_factories.go.tmpl", httpHandlerFactories))

//go:embed fixtures/workflow_factories.go.tmpl
var workflowFactories string
var WorkflowFactories = MustParse[codedef.FactoryContainer](DefaultOptions("workflow_factories.go.tmpl", workflowFactories))

//go:embed fixtures/activity_factories.go.tmpl
var activityFactories string
var ActivityFactories = MustParse[codedef.FactoryContainer](DefaultOptions("activity_factories.go.tmpl", activityFactories))
