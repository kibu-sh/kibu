package parser

import (
	"errors"
	"fmt"
	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"
	"go/ast"
	"go/types"
	"strconv"
)

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
		tags, _ := dir.Options.GetAll("tag", []string{"global"})
		orderOpt, _ := dir.Options.GetOne("order", "0")

		order, err := strconv.Atoi(orderOpt)
		if err != nil {
			err = errors.Join(err, fmt.Errorf("order must be an integer %s",
				meta.Position().String(),
			))
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
