// utils/random.go

package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GenerateHashUsername(provider string, socialID string) string {
	h := sha256.New()
	h.Write([]byte(provider + socialID))
	return fmt.Sprintf("u_%s", hex.EncodeToString(h.Sum(nil))[:10])
}
