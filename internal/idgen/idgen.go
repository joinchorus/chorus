package idgen

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// IDGenerator defines the interface for producing unique IDs.
type IDGenerator interface {
	GenerateID(prefix string) (string, error)
}

// RandomIDGenerator generates cryptographically secure random hex IDs with a prefix.
type RandomIDGenerator struct{}

// NewRandomIDGenerator constructs a RandomIDGenerator instance.
func NewRandomIDGenerator() *RandomIDGenerator {
	return &RandomIDGenerator{}
}

// GenerateID produces a random hex string ID prefixed with the given string.
func (g *RandomIDGenerator) GenerateID(prefix string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return prefix + hex.EncodeToString(bytes), nil
}
