package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEncryptionKey__String(t *testing.T) {
	var tests = map[string]struct {
		expected string
		key      EncryptionKey
	}{
		"should generate correct vault key": {
			expected: "hashivault://secret/data/kibu",
			key: EncryptionKey{
				Engine: "hashivault",
				Key:    "secret/data/kibu",
			},
		},

		"should generate correct gcp key": {
			expected: "gcpkms://secret/data/kibu",
			key: EncryptionKey{
				Engine: "gcpkms",
				Key:    "secret/data/kibu",
			},
		},
	}

	for s, s2 := range tests {
		t.Run(s, func(t *testing.T) {
			require.Equal(t, s2.expected, s2.key.String())
		})
	}
}
