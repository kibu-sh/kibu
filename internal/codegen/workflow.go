package codegen

import (
	_ "embed"
	"github.com/discernhq/devx/internal/codedef"
	"github.com/discernhq/devx/internal/codegen/templates"
)

//go:embed templates/fixtures/workflow.go.tmpl
var workflowTemplate string
var generateWorkflow = templates.MustParse[codedef.Module](
	templates.DefaultOptions("workflow", workflowTemplate),
)
