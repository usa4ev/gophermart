package argon2hash

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestComparePasswordAndHash(t *testing.T) {
	tests := []string{"qwerty12345", "aujlihf87993*(&#$&*HER>|", "ПарольКириллицей234"}

	for _, tt := range tests {

		t.Run("matching passwords", func(t *testing.T) {
			hash, err := GenerateFromPassword(tt, DefaultParams())
			require.NoError(t, err)

			ok, err := ComparePasswordAndHash(tt, hash)

			assert.Equal(t, true, ok, "passwords don't match when expected")
		})
	}

	for _, tt := range tests {

		t.Run("mismatching passwords", func(t *testing.T) {
			hash, err := GenerateFromPassword(tt, DefaultParams())
			require.NoError(t, err)

			ok, err := ComparePasswordAndHash(tt+tt[:3], hash)

			assert.Equal(t, false, ok, "passwords match when expected to not")
		})
	}
}
