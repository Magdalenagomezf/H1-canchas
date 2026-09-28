package mercadopago

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func VerifyWebhookSignature(secret, xSignature, xRequestID, dataID string) bool {
	ts, v1, ok := parseXSignature(xSignature)
	if !ok {
		return false
	}

	// La doc oficial de MP (ejemplos JS/Python/PHP) siempre pasa el data.id
	// a minúsculas antes de armar el manifest.
	manifest := "id:" + strings.ToLower(dataID) + ";request-id:" + xRequestID + ";ts:" + ts + ";"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(manifest))
	expected := hex.EncodeToString(mac.Sum(nil))

	// hmac.Equal (constant-time) because v1 is attacker-controlled input; a
	// plain == comparison would leak timing information about the secret.
	return hmac.Equal([]byte(expected), []byte(v1))
}

func parseXSignature(xSignature string) (ts, v1 string, ok bool) {
	for _, part := range strings.Split(xSignature, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}

		switch strings.TrimSpace(kv[0]) {
		case "ts":
			ts = strings.TrimSpace(kv[1])
		case "v1":
			v1 = strings.TrimSpace(kv[1])
		}
	}

	if ts == "" || v1 == "" {
		return "", "", false
	}

	return ts, v1, true
}
