package spec

import "errors"

var (
	ErrFileNotFoundInTxtArchive = errors.New("file not found in txt archive")
	ErrReadRoundTripTimeout     = errors.New("timed out waiting for round trip record")
	ErrSnapshotIDRequired       = errors.New("snapshot ID is required")
)
