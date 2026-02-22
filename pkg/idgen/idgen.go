package idgen

import (
	"crypto/rand"
	"encoding/hex"
)

func New() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}
