package random

import "math/rand/v2"

func NewRandomAlias(size int) string {
	alp := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	alias := make([]byte, size)
	for i := 0; i < size; i++ {
		alias[i] = alp[rand.IntN(len(alp))]
	}
	return string(alias)
}
