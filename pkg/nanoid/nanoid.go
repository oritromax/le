package nanoid

import (
	"crypto/rand"
	"math/big"
)


const idLetters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const defaultLen = 5

func New() string {
	return NewWithLen(defaultLen)
}

func NewWithLen(length int) string {
	result := make([]byte, length)
	for i := range result {
		randIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(idLetters))))
		result[i] = idLetters[randIdx.Int64()]
	}
	return string(result)
}
