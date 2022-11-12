package transport

// JSONDecoder decodes Request.Body into a concrete Endpoint request type.
//func JSONDecoder[T, R any](ctx context.Context, req Request, endpoint Endpoint[T, R]) (res R, err error) {
//	var body T
//	if err = json.NewDecoder(req.Body()).Decode(&body); err != nil {
//		return
//	}
//	return endpoint(ctx, body)
//}
