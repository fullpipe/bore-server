package utils

import "math/rand"

func RandString(n int, letterRunes []rune) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RandOTP() string {
	return RandString(6, []rune("0123456789"))
}
