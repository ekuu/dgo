package dgo

import (
	"context"
)

// Repo 仓储接口
type Repo[A AggBase] interface {
	Get(ctx context.Context, id ID) (A, error) // 当记录不存在时,应调用ErrAggregateNotFound()方法返回错误.
	List(ctx context.Context, ids ...ID) ([]A, error)
	Replay(ctx context.Context, id ID, version uint64) (*Snapshot[A], error)
	Save(ctx context.Context, a A, saveSnapshot bool) error
	Delete(ctx context.Context, a A) error
	SaveEvents(ctx context.Context, events Events) error
	Transaction(ctx context.Context, fn func(ctx context.Context, r Repo[A]) error) error
}
