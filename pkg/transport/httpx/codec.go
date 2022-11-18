package httpx

import "github.com/discernhq/devx/pkg/transport"

type Codec struct {
	transport.Encoder
	transport.Decoder
	transport.ErrorEncoder
}

var DefaultCodec = &Codec{
	Decoder:      DefaultDecoderChain(),
	Encoder:      JSONEncoder(),
	ErrorEncoder: JSONErrorEncoder(),
}
