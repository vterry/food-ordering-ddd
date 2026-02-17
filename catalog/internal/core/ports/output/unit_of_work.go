package output

import "context"

type UnitOfWork interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}
