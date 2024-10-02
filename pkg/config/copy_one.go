package config

import (
	"context"
	"github.com/pkg/errors"
)

func CopyOne(ctx context.Context, params CopyParams) (err error) {
	var data any
	if _, err = params.Source.GetByKey(ctx, params.SourcePath, &data); err != nil {
		err = errors.Wrapf(err, "failed to read source secret from %s", params.SourcePath)
		return
	}

	if _, err = params.Destination.Set(ctx, SetParams{
		Data:          data,
		Path:          params.SourcePath,
		EncryptionKey: params.DestinationKey,
	}); err != nil {
		err = errors.Wrapf(err, "failed to write destination secret to %s", params.SourcePath)
	}
	return
}
