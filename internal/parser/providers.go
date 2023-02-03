package parser

import (
	"github.com/discernhq/devx/internal/parser/directive"
	"go/token"
)

type Provider struct {
	Name       string
	File       *token.File
	Position   token.Position
	Directives directive.List
}

func collectProviders(p *Package) (err error) {
	for _, f := range p.funcIdCache.Keys() {
		ident, _ := p.funcIdCache.Get(f)
		dirs, ok := p.directiveCache.Get(ident)
		if !ok {
			return
		}

		if dirs.Some(directive.HasKey("devx", "provider")) {
			p.Providers.Set(ident, &Provider{
				Name:       f.Name(),
				Directives: dirs,
				File:       p.GoPackage.Fset.File(ident.Pos()),
				Position:   p.GoPackage.Fset.Position(ident.Pos()),
			})
		}
	}
	return
}
