package identity

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// IDGenerator defines the contract for generating unique identity IDs.
type IDGenerator interface {
	GenerateID() (string, error)
}

// RandomIDGenerator implements IDGenerator using cryptographically secure random bytes.
type RandomIDGenerator struct{}

// NewRandomIDGenerator creates a new RandomIDGenerator.
func NewRandomIDGenerator() *RandomIDGenerator {
	return &RandomIDGenerator{}
}

// GenerateID produces a random 16-byte hex string ID with a prefix.
func (g *RandomIDGenerator) GenerateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random ID: %w", err)
	}
	return "usr_" + hex.EncodeToString(bytes), nil
}
