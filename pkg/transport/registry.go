package transport

import orderedmap "github.com/wk8/go-ordered-map/v2"

type EndpointType string

const (
	EndpointTypeHTTP             EndpointType = "http"
	EndpointTypeGRPC             EndpointType = "grpc"
	EndpointTypeWS               EndpointType = "ws"
	EndpointTypeGraphQL          EndpointType = "graphql"
	EndpointTypeNats             EndpointType = "nats"
	EndpointTypeTemporalActivity EndpointType = "temporal.activity"
	EndpointTypeTemporalWorkflow EndpointType = "temporal.workflow"
)

type EndpointRegistry interface {
	Register(EndpointInfo)
}

type EndpointMetadata interface {
	EndpointMeta()
}

type EndpointInfo struct {
	Type     EndpointType
	Package  string
	Service  string
	Method   string
	Tags     []string
	Handler  Handler
	Metadata EndpointMetadata
}

func NewEndpointInfo() EndpointInfo {
	return EndpointInfo{}
}

func (e EndpointInfo) Key() string {
	return e.Package + "." + e.Service + "." + e.Method
}

func (e EndpointInfo) WithTags(tags ...string) EndpointInfo {
	e.Tags = append(e.Tags, tags...)
	return e
}

func (e EndpointInfo) WithHandler(handler Handler) EndpointInfo {
	e.Handler = handler
	return e
}

func (e EndpointInfo) WithType(t EndpointType) EndpointInfo {
	e.Type = t
	return e
}

func (e EndpointInfo) WithPackage(pkg string) EndpointInfo {
	e.Package = pkg
	return e
}

func (e EndpointInfo) WithService(service string) EndpointInfo {
	e.Service = service
	return e
}

func (e EndpointInfo) WithMethod(method string) EndpointInfo {
	e.Method = method
	return e
}

var _ EndpointRegistry = (*endpointRegistry)(nil)

type endpointRegistry struct {
	cache *orderedmap.OrderedMap[string, EndpointInfo]
}

func (e endpointRegistry) Register(info EndpointInfo) {
	e.cache.Set(info.Key(), info)
}

func NewRegistry() EndpointRegistry {
	return &endpointRegistry{
		cache: orderedmap.New[string, EndpointInfo](),
	}
}

type HTTPMetadata struct {
	Path   string
	Method string
}

func (m HTTPMetadata) EndpointMeta() {}
