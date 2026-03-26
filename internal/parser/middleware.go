package parser

import (
	"errors"
	"go/ast"
	"go/types"
	"strconv"

	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"
)

var ErrMiddlewareOrderNotInteger = errors.New("middleware order must be an integer")

type Middleware struct {
	*TypeMeta
	Name       string
	Tags       []string
	Order      int
	Directives decorators.List
}

func collectMiddleware(p *Package) defMapperFunc {
	return func(ident *ast.Ident, obj types.Object) (err error) {
		dirs, ok := p.directiveCache[ident]
		if !ok {
			return
		}

		dir, isMiddleware := dirs.Find(decorators.HasKey("kibu", "middleware"))
		if !isMiddleware {
			return
		}

		meta := NewTypeMeta(ident, obj, p)
		tags, _ := dir.Options.ListValues("tag", []string{"global"})
		orderOpt := dir.Options.Lookup("order").Or("0")

		order, err := strconv.Atoi(orderOpt)
		if err != nil {
			err = errors.Join(NewPositionError(ErrMiddlewareOrderNotInteger, meta.Position()), err)
			return
		}

		mw := &Middleware{
			Name:       ident.Name,
			Directives: dirs,
			Order:      order,
			Tags:       tags,
			TypeMeta:   meta,
		}

		p.Middleware[ident] = mw
		return
	}
}
