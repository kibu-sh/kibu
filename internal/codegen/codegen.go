package codegen

import (
	"bytes"
	"github.com/discernhq/devx/internal/codedef"
	"github.com/discernhq/devx/internal/codegen/templates"
	"os"
	"path/filepath"
)

type Generator struct {
	Clobber  bool
	Executor templates.ExecFunc[codedef.Module]
}

type Pipeline struct {
	Generators map[string]Generator
}

func (p Pipeline) Generate(dir string, mod codedef.Module) (err error) {
	for f, g := range p.Generators {
		var data bytes.Buffer
		path := filepath.Join(dir, f)

		if _, err = os.Stat(path); err == nil && !g.Clobber {
			continue
		}

		data, err = g.Executor(mod)
		if err != nil {
			return
		}
		if err = os.WriteFile(path, data.Bytes(), 0744); err != nil {
			return
		}
	}
	return
}

var DefaultPipeline = Pipeline{
	Generators: map[string]Generator{
		"devx.gen.go": {
			Clobber:  true,
			Executor: templates.DevxGen,
		},
		"workflow.go": {
			Executor: templates.Workflow,
		},
		"activity.go": {
			Executor: templates.Activity,
		},
	},
}
