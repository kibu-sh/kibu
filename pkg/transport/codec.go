package transport

type Codec interface {
	Decoder
	Encoder
	ErrorEncoder
}
