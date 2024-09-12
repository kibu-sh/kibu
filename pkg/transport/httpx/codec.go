package httpx

import "github.com/kibu-sh/kibu/pkg/transport"

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
