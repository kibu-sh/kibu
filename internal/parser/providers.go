package parser

import (
	"github.com/kibu-sh/kibu/internal/parser/directive"
	"github.com/pkg/errors"
	"go/ast"
	"go/types"
)

type ProviderType string

var (
	StructProviderType   ProviderType = "struct"
	FunctionProviderType ProviderType = "function"
)

type Provider struct {
	*TypeMeta
	Name       string
	Type       ProviderType
	Directives directive.List
}

func collectProviders(p *Package) defMapperFunc {
	return func(ident *ast.Ident, obj types.Object) (err error) {
		dirs, ok := p.directiveCache[ident]
		if !ok {
			return
		}

		if !dirs.Some(directive.HasKey("kibue", "provider")) {
			return
		}

		prv := &Provider{
			Name:       ident.Name,
			Directives: dirs,
			TypeMeta:   NewTypeMeta(ident, obj, p),
		}

		switch obj.Type().(type) {
		case *types.Named:
			prv.Type = StructProviderType
		case *types.Signature:
			prv.Type = FunctionProviderType
		default:
			err = errors.Errorf("unsupported provider type: %s at %s",
				obj.Type().String(),
				prv.Position().String(),
			)
			return
		}

		p.Providers[ident] = prv
		return
	}
}
