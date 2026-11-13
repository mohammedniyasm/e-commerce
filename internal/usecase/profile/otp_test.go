package profile

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileOTPHelpers(t *testing.T) {
	t.Run("generateOTP produces 6-digit numeric string", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			otp, err := generateOTP()
			require.NoError(t, err)
			assert.Len(t, otp, 6)
			for _, ch := range otp {
				assert.GreaterOrEqual(t, ch, '0')
				assert.LessOrEqual(t, ch, '9')
			}
		}
	})

	t.Run("email change redis key helpers", func(t *testing.T) {
		otpKey := emailChangeOTPKey(42)
		pendingKey := emailChangePendingEmailKey(42)
		assert.Contains(t, otpKey, "42")
		assert.Contains(t, pendingKey, "42")
		assert.NotEqual(t, otpKey, pendingKey)
	})
}
