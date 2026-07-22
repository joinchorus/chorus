package thread

import "context"

// Repository defines storage operations for threads and messages.
type Repository interface {
	SaveThread(ctx context.Context, t *Thread) error
	FindThreadByID(ctx context.Context, id string) (*Thread, error)
	ListThreads(ctx context.Context) ([]*Thread, error)

	SaveMessage(ctx context.Context, m *Message) error
	ListMessagesByThreadID(ctx context.Context, threadID string) ([]*Message, error)
}
