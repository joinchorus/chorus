package identity

import "time"

// Identity represents a user or actor identity in the system.
// Serves as the Single Source of Truth (SSOT) model for identity domain & JSON representation.
type Identity struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateInput holds data required to create a new Identity.
type CreateInput struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}
