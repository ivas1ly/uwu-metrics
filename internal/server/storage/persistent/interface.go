package persistent

import "context"

type Storage interface {
	Save(ctx context.Context) error
	Restore(ctx context.Context) error
}
