package utils

import (
	"math/rand"
	"time"
)

func GenerateRandom(stringLength int) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	letters := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, stringLength)
	for i := range b {
		b[i] = letters[random.Intn(len(letters))]
	}

	return string(b)
}
