package config

import "context"

type CopyParams struct {
	Source         Store
	SourcePath     string
	Destination    Store
	DestinationKey EncryptionKey
}

type CopyFunc func(ctx context.Context, params CopyParams) (err error)

func CopyRecursive(ctx context.Context, params CopyParams) (err error) {
	iter, err := params.Source.List(ctx, ListParams{
		Path: params.SourcePath,
	})

	if err != nil {
		return err
	}

	for item := range iter.Next() {
		if err = item.Error(); err != nil {
			return
		}

		var data any
		if _, err = item.Get(ctx, &data); err != nil {
			return
		}

		_, err = params.Destination.Set(ctx, SetParams{
			Data:          data,
			Path:          item.Path(),
			EncryptionKey: params.DestinationKey,
		})
		if err != nil {
			return err
		}
	}

	return
}
