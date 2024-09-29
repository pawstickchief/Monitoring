package clienttoken

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

func GenerateAuthToken(ip string) string {
	hash := sha256.New()
	hash.Write([]byte(ip + time.Now().String()))
	return hex.EncodeToString(hash.Sum(nil))
}
