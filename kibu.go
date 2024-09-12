package kibue

// issues with type aliases
// https://github.com/golang/go/milestone/250
// https://github.com/golang/go/issues/46477#issuecomment-1134888278
// type MiddlewareFunc[T any] func(T) T
//
// func WithMiddleware[T any](base T, middleware ...MiddlewareFunc[T]) T {
// 	for _, m := range middleware {
// 		base = m(base)
// 	}
// 	return base
// }
