package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Calc(key string, payload []byte) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(payload)
	hashSum := h.Sum(nil)

	return hex.EncodeToString(hashSum)
}
