package mercadopago

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func computeSignature(t *testing.T, secret, manifest string) string {
	t.Helper()

	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(manifest)); err != nil {
		t.Fatalf("mac.Write() error = %v", err)
	}

	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookSignature(t *testing.T) {
	const (
		secret     = "test-webhook-secret"
		dataID     = "123456789"
		xRequestID = "req-abc-123"
		ts         = "1700000000000"
	)

	manifest := "id:" + dataID + ";request-id:" + xRequestID + ";ts:" + ts + ";"
	validDigest := computeSignature(t, secret, manifest)
	validXSignature := "ts=" + ts + ",v1=" + validDigest

	tests := []struct {
		name       string
		secret     string
		xSignature string
		xRequestID string
		dataID     string
		want       bool
	}{
		{
			name:       "valid signature returns true",
			secret:     secret,
			xSignature: validXSignature,
			xRequestID: xRequestID,
			dataID:     dataID,
			want:       true,
		},
		{
			name:       "v1 before ts still parses and validates",
			secret:     secret,
			xSignature: "v1=" + validDigest + ",ts=" + ts,
			xRequestID: xRequestID,
			dataID:     dataID,
			want:       true,
		},
		{
			name:       "wrong secret returns false",
			secret:     "another-secret",
			xSignature: validXSignature,
			xRequestID: xRequestID,
			dataID:     dataID,
			want:       false,
		},
		{
			name:       "tampered dataID returns false",
			secret:     secret,
			xSignature: validXSignature,
			xRequestID: xRequestID,
			dataID:     "999999999",
			want:       false,
		},
		{
			name:       "missing v1 returns false",
			secret:     secret,
			xSignature: "ts=" + ts,
			xRequestID: xRequestID,
			dataID:     dataID,
			want:       false,
		},
		{
			name:       "missing ts returns false",
			secret:     secret,
			xSignature: "v1=" + validDigest,
			xRequestID: xRequestID,
			dataID:     dataID,
			want:       false,
		},
		{
			name:       "empty header returns false",
			secret:     secret,
			xSignature: "",
			xRequestID: xRequestID,
			dataID:     dataID,
			want:       false,
		},
		{
			name:       "malformed header without key=value pairs returns false",
			secret:     secret,
			xSignature: "not-a-valid-header",
			xRequestID: xRequestID,
			dataID:     dataID,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyWebhookSignature(tt.secret, tt.xSignature, tt.xRequestID, tt.dataID)
			if got != tt.want {
				t.Errorf("VerifyWebhookSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}
