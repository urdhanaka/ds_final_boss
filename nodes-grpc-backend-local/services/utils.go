package services

import (
	rand "math/rand"
	randv2 "math/rand/v2"
	"time"
)

// get random number [0,maxNum)
func getRandomIndex(maxNum int) int {
	return randv2.IntN(maxNum)
}

func generateRandom(stringLength int) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	letters := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, stringLength)
	for i := range b {
		b[i] = letters[random.Intn(len(letters))]
	}

	return string(b)
}
