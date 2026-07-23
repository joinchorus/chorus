package thread

import "time"

// Thread represents a discussion thread.
type Thread struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	AuthorID         string    `json:"author_id,omitempty"`
	ConversationName string    `json:"conversation_name"`
	Country          *string   `json:"country"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Message represents a post or reply within a thread.
type Message struct {
	ID               string    `json:"id"`
	ThreadID         string    `json:"thread_id"`
	AuthorID         string    `json:"author_id,omitempty"`
	ConversationName string    `json:"conversation_name"`
	Country          *string   `json:"country"`
	Content          string    `json:"content"`
	CreatedAt        time.Time `json:"created_at"`
}

// ThreadDetail contains thread header metadata and all posted messages.
type ThreadDetail struct {
	Thread   *Thread    `json:"thread"`
	Messages []*Message `json:"messages"`
}

// CreateThreadInput holds parameters for initializing a new thread.
type CreateThreadInput struct {
	Title            string `json:"title"`
	Body             string `json:"body"`
	ShowCountry      bool   `json:"show_country"`
	ConversationName string `json:"conversation_name,omitempty"`
}

// CreateMessageInput holds parameters for appending a reply to a thread.
type CreateMessageInput struct {
	Body             string `json:"body"`
	Content          string `json:"content"` // Alias fallback for legacy content key
	ShowCountry      bool   `json:"show_country"`
	ConversationName string `json:"conversation_name,omitempty"`
}

// TranslateMessageInput holds parameters for requesting on-demand translation.
type TranslateMessageInput struct {
	TargetLang string `json:"target_lang"`
}

// GetBody returns Body or Content field.
func (i *CreateMessageInput) GetBody() string {
	if i.Body != "" {
		return i.Body
	}
	return i.Content
}

