package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenVerify(t *testing.T) {
	type args struct {
		userID   string
		lifeTime time.Duration
	}

	tt := args{userID: "user1", lifeTime: 5 * time.Second}

	t.Run("open/close no error", func(t *testing.T) {
		got, _, err := Open(tt.userID, tt.lifeTime)
		require.NoError(t, err)

		userID, err := Verify(got)
		require.NoError(t, err)
		assert.Equal(t, tt.userID, userID)
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := Verify("not a token")
		if err == nil {
			t.Errorf("invalid token passed validation")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		got, _, err := Open(tt.userID, tt.lifeTime)
		require.NoError(t, err)

		timer := time.NewTimer(6 * time.Second)

		<-timer.C

		_, err = Verify(got)
		if err == nil {
			t.Errorf("expired token passed validation")
		}
	})
}
