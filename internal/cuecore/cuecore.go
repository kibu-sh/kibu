package cuecore

import (
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/discernhq/devx/internal/codedef"
)

type LoadOptions struct {
	Dir        string
	Entrypoint []string
}

func Load(opts LoadOptions) (mod codedef.Module, err error) {
	ctx := cuecontext.New()
	instances := load.Instances(opts.Entrypoint, &load.Config{
		Dir: opts.Dir,
		//Package: "_",
		//DataFiles:   false,
		//Overlay:     nil,
		//Stdin:       nil,
	})

	values, err := ctx.BuildInstances(instances)
	if err != nil {
		return
	}

	err = values[0].Validate()
	if err != nil {
		return
	}

	mod.Name = instances[0].PkgName

	err = values[0].Decode(&mod)
	if err != nil {
		return
	}
	return
}
