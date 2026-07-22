package identity

import "context"

// Repository defines storage operations for Identity.
type Repository interface {
	Save(ctx context.Context, id *Identity) error
	FindByID(ctx context.Context, id string) (*Identity, error)
	FindByEmail(ctx context.Context, email string) (*Identity, error)
}
