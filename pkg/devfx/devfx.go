package devfx

import (
	"fmt"
	"github.com/fatih/structtag"
	"github.com/kibu-sh/kibu/pkg/transport"
	"github.com/kibu-sh/kibu/pkg/transport/httpx"
	"github.com/samber/lo"
	"go.uber.org/fx"
)

type ServiceKey struct {
	Pkg  string
	Name string
}

func (sk ServiceKey) String() string {
	return fmt.Sprintf(`pkg:%s:service:%s`, sk.Pkg, sk.Name)
}

func (sk ServiceKey) NameTag() structtag.Tag {
	return structtag.Tag{
		Key:  "name",
		Name: sk.String(),
	}
}

type EndpointKey struct {
	ServiceKey ServiceKey
	Name       string
}

func (ek EndpointKey) String() string {
	return fmt.Sprintf(`%s:endpoint:%s`, ek.ServiceKey.String(), ek.Name)
}

func (ek EndpointKey) NameTag() structtag.Tag {
	return structtag.Tag{
		Key:  "name",
		Name: ek.String(),
	}
}

func (ek EndpointKey) GroupTag() structtag.Tag {
	return structtag.Tag{
		Key:  "group",
		Name: ek.String(),
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

func AsMiddleware(provider any, endpointKey EndpointKey) fx.Option {
	return fx.Provide(
		fx.Annotate(
			provider,
			ResultTags(endpointKey.GroupTag()),
		),
	)
}

func HandlerGroupTag() structtag.Tag {
	return structtag.Tag{
		Key: "group",
		// TODO: think about adding suffixes to group tags for kibue
		// kibue:controllers
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

type EndpointProviderFunc[Service any] func(service *Service, middleware []transport.Middleware) transport.Handler

func AsEndpoint[Service any](
	provider EndpointProviderFunc[Service],
	serviceKey ServiceKey,
	endpointKey EndpointKey,
) fx.Option {
	return fx.Provide(
		fx.Annotate(
			provider,
			ParamTags(serviceKey.NameTag(), endpointKey.GroupTag()),
			ResultTags(endpointKey.NameTag()),
		),
	)
}

type HTTPProviderFunc func(endpoint transport.Handler) *httpx.Handler

func AsHTTPEndpoint(
	provider HTTPProviderFunc,
	endpointKey EndpointKey,
	handlerGroupTag structtag.Tag,
) fx.Option {
	return fx.Provide(
		fx.Annotate(
			provider,
			ParamTags(endpointKey.NameTag()),
			ResultTags(handlerGroupTag),
		),
	)
}
