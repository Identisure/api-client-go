package idkollen

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// VerifyCallbackSignature verifies the X-Signature-SHA256 header of a callback request.
// body must be the raw request body as received; verify before parsing and never re-serialize.
// It returns false for a missing, malformed or mismatching signature.
func VerifyCallbackSignature(body []byte, signature, secret string) bool {
	expected, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	return hmac.Equal(mac.Sum(nil), expected)
}
