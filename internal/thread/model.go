package thread

import "time"

// Thread represents a discussion or messaging thread.
type Thread struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	AuthorID  string    `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message represents an individual message posted within a thread.
type Message struct {
	ID        string    `json:"id"`
	ThreadID  string    `json:"thread_id"`
	AuthorID  string    `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateThreadInput holds data required to initialize a thread.
type CreateThreadInput struct {
	Title    string `json:"title"`
	AuthorID string `json:"author_id"`
}

// CreateMessageInput holds data required to append a message to a thread.
type CreateMessageInput struct {
	AuthorID string `json:"author_id"`
	Content  string `json:"content"`
}
