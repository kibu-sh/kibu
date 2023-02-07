package parser

import (
	"github.com/discernhq/devx/internal/parser/directive"
	"github.com/pkg/errors"
	"go/ast"
	"go/token"
	"go/types"
)

type ProviderType string

var (
	StructProviderType   ProviderType = "struct"
	FunctionProviderType ProviderType = "function"
)

type Provider struct {
	Name       string
	Type       ProviderType
	File       *token.File
	Position   token.Position
	Directives directive.List
}

func collectProviders(p *Package) defMapperFunc {
	return func(ident *ast.Ident, obj types.Object) (err error) {
		dirs, ok := p.directiveCache.Get(ident)
		if !ok {
			return
		}

		if !dirs.Some(directive.HasKey("devx", "provider")) {
			return
		}

		prv := &Provider{
			Name:       ident.Name,
			Directives: dirs,
			File:       p.GoPackage.Fset.File(obj.Pos()),
			Position:   p.GoPackage.Fset.Position(obj.Pos()),
		}

		switch obj.Type().(type) {
		case *types.Named:
			prv.Type = StructProviderType
		case *types.Signature:
			prv.Type = FunctionProviderType
		default:
			err = errors.Errorf("unsupported provider type: %s at %s:%d",
				obj.Type().String(),
				prv.Position.Filename,
				prv.Position.Line,
			)
			return
		}

		p.Providers.Set(ident, prv)
		return
	}
}
