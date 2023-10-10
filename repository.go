package dgo

import (
	"context"
)

// Repo 仓储接口
type Repo[A AggBase] interface {
	// Get 当记录不存在时,应调用返回ErrNotFound.
	Get(ctx context.Context, id ID) (A, error)
	List(ctx context.Context, ids ...ID) ([]A, error)
	Save(ctx context.Context, a A) error
	Delete(ctx context.Context, a A) error
	SaveEvents(ctx context.Context, events Events) error
	Transaction(ctx context.Context, fn func(ctx context.Context, r Repo[A]) error) error
}
