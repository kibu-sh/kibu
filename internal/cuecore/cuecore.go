package cuecore

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"errors"
)

type Loader struct {
	Dir        string
	Entrypoint []string

	Ctx       *cue.Context
	Instances []*build.Instance
	Values    []cue.Value
}

type LoadOption func(l *Loader) error

func WithDir(dir string) LoadOption {
	return func(l *Loader) error {
		l.Dir = dir
		return nil
	}
}

func WithEntrypoint(entrypoint []string) LoadOption {
	return func(l *Loader) error {
		l.Entrypoint = entrypoint
		return nil
	}
}

func WithCueContext(ctx *cue.Context) LoadOption {
	return func(l *Loader) error {
		l.Ctx = ctx
		return nil
	}
}

func WithValidation() LoadOption {
	return func(l *Loader) (err error) {
		for _, value := range l.Values {
			err = value.Validate()
			if err != nil {
				return
			}
		}
		return
	}
}

func WithBasicDecoder[T any](v *T) LoadOption {
	return func(l *Loader) error {
		if len(l.Values) == 0 {
			return errors.New("no cue values to decode, did you build a cue instance")
		}
		return l.Values[0].Decode(v)
	}
}

func WithBuild() LoadOption {
	return func(l *Loader) (err error) {
		l.Instances = load.Instances(l.Entrypoint, &load.Config{
			Dir: l.Dir,
			// Package: "_",
			// DataFiles:   false,
			// Overlay:     nil,
			// Stdin:       nil,
		})

		l.Values, err = l.Ctx.BuildInstances(l.Instances)
		if err != nil {
			return
		}
		return
	}
}

func Load(opts ...LoadOption) (l *Loader, err error) {
	l = new(Loader)
	for _, opt := range opts {
		if err = opt(l); err != nil {
			return
		}
	}
	return
}

func LoadWithDefaults(dir string, entrypoint []string, opts ...LoadOption) (l *Loader, err error) {
	opts = append([]LoadOption{
		WithDir(dir),
		WithEntrypoint(entrypoint),
		WithCueContext(cuecontext.New()),
		WithBuild(),
	}, opts...)

	return Load(opts...)
}
