package main

import (
	"crypto/rand"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const codeLength = 6

func generateShortCode() (string, error) {

	short_code := make([]byte, codeLength)

	random_bytes := make([]byte, codeLength)

	_, err := rand.Read(random_bytes)
	if err != nil {
		return "", err
	}

	//Convert each random byte to a character from our charset
	for i := 0; i < codeLength; i++ {
		// Use modulo to get index in range 0-61
		index := random_bytes[i] % 62
		short_code[i] = charset[index]
	}

	return string(short_code), nil
}
