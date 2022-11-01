package codegen

import (
	_ "embed"
	"github.com/discernhq/devx/internal/codedef"
	"github.com/discernhq/devx/internal/codegen/templates"
)

//go:embed templates/fixtures/activity.go.tmpl
var activityTemplate string
var generateActivity = templates.MustParse[codedef.Module](
	templates.DefaultOptions("activity", activityTemplate),
)
