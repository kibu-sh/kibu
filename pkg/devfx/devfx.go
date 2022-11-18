package devfx

import (
	"fmt"
	"github.com/discernhq/devx/pkg/transport"
	"github.com/discernhq/devx/pkg/transport/httpx"
	"github.com/fatih/structtag"
	"github.com/samber/lo"
	"go.uber.org/fx"
)

type ServiceKey struct {
	Pkg  string
	Name string
}

func (sk ServiceKey) String() string {
	return fmt.Sprintf(`%s:%s`, sk.Pkg, sk.Name)
}

func (sk ServiceKey) NameTag() structtag.Tag {
	return structtag.Tag{
		Key:  "name",
		Name: sk.String(),
	}
}

func (sk ServiceKey) MethodString(method string) string {
	return fmt.Sprintf("%s:%s", sk.String(), method)
}

func (sk ServiceKey) MethodTag(s string) structtag.Tag {
	return structtag.Tag{
		Key:     "group",
		Name:    sk.MethodString(s),
		Options: nil,
	}
}

func AsService(provider any, sk ServiceKey) fx.Option {
	return fx.Provide(
		fx.Annotate(
			provider,
			ResultTags(sk.NameTag()),
		),
	)
}

func AsMiddleware(provider any, tag structtag.Tag) fx.Option {
	return fx.Provide(
		fx.Annotate(
			provider,
			ResultTags(tag),
		),
	)
}

type EndpointProvider[T any] func(service *T, middleware []transport.Middleware) *httpx.Handler

func BindServiceEndpoint[T any](
	provider EndpointProvider[T],
	serviceTag structtag.Tag,
	methodTag structtag.Tag,
) fx.Option {
	return fx.Provide(
		fx.Annotate(
			provider,
			ParamTags(serviceTag, methodTag),
			ResultTags(HandlerGroupTag()),
		),
	)
}

func AsEndpoint(provider any, tag structtag.Tag) fx.Option {
	return fx.Provide(
		fx.Annotate(
			provider,
			ParamTags(tag),
			ResultTags(HandlerGroupTag()),
		),
	)
}

func AsHandler(provider any) fx.Option {
	return fx.Provide(
		fx.Annotate(
			provider,
			fx.As(new(httpx.Handler)),
			fx.ResultTags(`group:"handlers"`),
		),
	)
}

func HandlerGroupTag() structtag.Tag {
	return structtag.Tag{
		Key: "group",
		// TODO: think about adding suffixes to group tags for devx
		// devx:controllers
		Name: "handlers",
	}
}

func ResultTags(tags ...structtag.Tag) fx.Annotation {
	return fx.ResultTags(
		lo.Map(tags, func(t structtag.Tag, _ int) string {
			return t.String()
		})...,
	)
}

func ParamTags(tags ...structtag.Tag) fx.Annotation {
	return fx.ParamTags(
		lo.Map(tags, func(t structtag.Tag, _ int) string {
			return t.String()
		})...,
	)
}
