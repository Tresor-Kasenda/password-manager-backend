package main

import (
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

func main() {
	// Stored hash and salt from database
	hashStr := "0YDvAmOKe014YuIMI1LfV8+3VjqhkuLe7VpN2jXStzQ="
	saltStr := "RfeEQs/aCt2ocS6IiA1dW9dMFrA8t0LEbLAECA17SZg="
	password := "password123"

	// Decode hash and salt
	hash, _ := base64.StdEncoding.DecodeString(hashStr)
	salt, _ := base64.StdEncoding.DecodeString(saltStr)

	// Derive key with same parameters as in crypto.go
	derivedKey := argon2.IDKey(
		[]byte(password),
		salt,
		3,
		64*1024,
		4,
		32,
	)

	fmt.Printf("Stored hash:   %s\n", hashStr)
	fmt.Printf("Derived hash:  %s\n", base64.StdEncoding.EncodeToString(derivedKey))
	fmt.Printf("Match: %v\n", string(hash) == string(derivedKey))
}
