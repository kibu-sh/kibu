package a

import "github.com/google/wire"

//kibu:provider import=github.com/kibu-sh/kibu/pkg/transport/httpx group=HandlerFactory
type Service struct{}

//kibu:provider
func ProviderFunc() Service {
	return Service{}
}

//kibu:provider
var WireSet = wire.NewSet(
	ProviderFunc,
	wire.Struct(new(Service), "*"),
)
