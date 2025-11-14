package utils

import (
	"crypto/sha256"
	"fmt"
)

func HashIP(ip string) string {
	hash := sha256.Sum256([]byte(ip))
	return fmt.Sprintf("%x", hash)
}
