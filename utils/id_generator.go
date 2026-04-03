package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// GeneratePublicRoomRequestID generates a unique, shareable ID for room requests
// Format: REQ-YYYYMMDD-XXXX (where XXXX is 4 random alphanumeric characters)
func GeneratePublicRoomRequestID() string {
	timestamp := time.Now().Format("20060102")
	const allowedChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 4)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(allowedChars))))
		b[i] = allowedChars[n.Int64()]
	}
	return fmt.Sprintf("REQ-%s-%s", timestamp, string(b))
}
