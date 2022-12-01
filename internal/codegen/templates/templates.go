package templates

import (
	_ "embed"
	"github.com/discernhq/devx/internal/codedef"
)

//go:embed fixtures/devx.gen.go.tmpl
var devxGen string
var DevxGen = MustParse[codedef.Module](DefaultOptions("devx.gen.go.tmpl", devxGen))

//go:embed fixtures/activity.go.tmpl
var activity string
var Activity = MustParse[codedef.Module](DefaultOptions("activity.go.tmpl", activity))

//go:embed fixtures/workflow.go.tmpl
var workflow string
var Workflow = MustParse[codedef.Module](DefaultOptions("workflow.go.tmpl", workflow))

//go:embed fixtures/service.go.tmpl
var service string
var Service = MustParse[codedef.Module](DefaultOptions("service.go.tmpl", service))
