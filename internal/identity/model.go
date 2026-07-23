package identity

import "time"

// Identity represents a user or actor identity in the system.
// Serves as the Single Source of Truth (SSOT) model for identity domain & JSON representation.
type Identity struct {
	ID               string    `json:"id,omitempty"`
	ConversationName string    `json:"conversation_name"`
	Country          *string   `json:"country,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// CreateInput holds data required to create a new Identity.
type CreateInput struct {
	ConversationName string `json:"conversation_name,omitempty"`
	ShowCountry      bool   `json:"show_country,omitempty"`
}
