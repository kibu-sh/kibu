package kwarg

type Option[O any] func(*O)
type OptionWithErr[O any] func(*O) error

func BuildOptions[O any](opts ...Option[O]) (o O) {
	for _, opt := range opts {
		opt(&o)
	}
	return
}
func BuildOptionsWithErr[O any](opts ...OptionWithErr[O]) (o O, err error) {
	for _, opt := range opts {
		if err = opt(&o); err != nil {
			return
		}
	}
	return
}
