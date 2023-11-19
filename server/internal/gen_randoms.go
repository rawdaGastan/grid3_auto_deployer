// Package internal for internal details
package internal

import (
	"math/rand"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// GenerateRandomVoucher generates a random voucher
func GenerateRandomVoucher(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// GenerateRandomCode generates random code of 4 digits
func GenerateRandomCode() int {
	min := 1000
	max := 9999
	return rand.Intn(max-min) + min
}
