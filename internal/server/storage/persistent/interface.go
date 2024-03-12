package persistent

import "context"

// Storage is the interface that groups the persistent storage methods.
type Storage interface {
	Save(ctx context.Context) error
	Restore(ctx context.Context) error
}
