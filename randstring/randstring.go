// See Icza's answer to "How to generate a random string of a fixed
// length in Go?" at https://stackoverflow.com/a/31832326 for a
// discussion on simplicity and performance

package randstring

import (
	"math/rand"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandString produces a randomly generated string of length n
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
