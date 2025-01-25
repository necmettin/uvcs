package utils

import (
	"math/rand"
	"time"
)

func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
}

// GenerateRandomInt generates a random integer
func GenerateRandomInt() int {
	return rand.Intn(1000000)
}
