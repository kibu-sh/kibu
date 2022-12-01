package build

import (
	"github.com/discernhq/devx/internal/codedef"
	"github.com/discernhq/devx/internal/codegen"
	"github.com/discernhq/devx/internal/cuecore"
	"path/filepath"
)

type Options func(b *Builder) error

func WithDir(dir string) Options {
	return func(b *Builder) error {
		b.Dir = dir
		return nil
	}
}

func WithEntrypoint(entrypoint string) Options {
	return func(b *Builder) error {
		if !filepath.IsAbs(entrypoint) {
			entrypoint = filepath.Join(b.Dir, entrypoint)
		}
		return nil
	}
}

type Builder struct {
	Dir        string
	Entrypoint []string
}

func NewBuilder(opts ...Options) (b *Builder, err error) {
	b = new(Builder)
	for _, opt := range opts {
		if err = opt(b); err != nil {
			return
		}
	}
	return
}

func NewWithDefaults(dir string) (b *Builder, err error) {
	return NewBuilder(WithDir(dir), WithEntrypoint("devx.module.cue"))
}

func (b *Builder) Exec() (err error) {
	loader, err := cuecore.NewDefaultLoader(b.Dir, b.Entrypoint)
	if err != nil {
		return
	}

	var mod codedef.Module
	err = loader.Load(cuecore.WithValidation(), cuecore.WithBasicDecoder(&mod))

	if err != nil {
		return
	}

	mod.Name = instances[0].PkgName

	err = codegen.DefaultPipeline.Generate(b.Dir, mod)
	if err != nil {
		return
	}

	err = reconcileDrift(b.Dir, mod)
	if err != nil {
		return
	}

	return
}
