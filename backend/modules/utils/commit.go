package utils

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// GenerateCommitHash generates a unique commit hash using timestamp and random data
func GenerateCommitHash() string {
	timestamp := time.Now().UnixNano()
	randomKey := GenerateSecurityKey()
	data := fmt.Sprintf("%d-%s", timestamp, randomKey)

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)[:40] // Return first 40 characters for git-like hash
}
