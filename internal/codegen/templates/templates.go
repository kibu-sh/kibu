package templates

import (
	_ "embed"
	"github.com/discernhq/devx/internal/codedef"
)

//go:embed fixtures/devx.gen.go.tmpl
var devxGen string
var DevxGen = MustParse[codedef.Module](DefaultOptions("magnet", devxGen))

//go:embed fixtures/activity.go.tmpl
var activity string
var Activity = MustParse[codedef.Module](DefaultOptions("activity", activity))

//go:embed fixtures/workflow.go.tmpl
var workflow string
var Workflow = MustParse[codedef.Module](DefaultOptions("workflow", workflow))
