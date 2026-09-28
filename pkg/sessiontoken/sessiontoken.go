// Package sessiontoken creates opaque session tokens and the digest that is
// the only form of a token the database ever stores.
package sessiontoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// New returns a random 256-bit token, hex encoded. It is handed to the client
// once and never persisted.
func New() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b) // crypto/rand.Read never returns an error (Go 1.24+)
	return hex.EncodeToString(b)
}

// Hash returns the hex SHA-256 of token. Migration 000011 hashes existing rows
// with the same function (encode(sha256(convert_to(token, 'UTF8')), 'hex')),
// so the two must stay in sync.
func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
