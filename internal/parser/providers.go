package parser

import "github.com/discernhq/devx/internal/parser/directive"

type Provider struct {
	Name       string
	Directives directive.List
}

func collectProviders(p *Package) (err error) {
	for f, ident := range p.funcIdCache {
		dirs, ok := p.directiveCache[ident]
		if !ok {
			return
		}

		if dirs.Some(directive.HasKey("devx", "provider")) {
			p.Providers[ident] = &Provider{
				Name:       f.Name(),
				Directives: dirs,
			}
		}
	}
	return
}
