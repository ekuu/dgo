package dgo

import (
	"context"
	"log/slog"
	"reflect"

	"github.com/ekuu/dgo/internal"
	clone "github.com/huandu/go-clone/generic"
	pkgerr "github.com/pkg/errors"
)

// Handler 命令处理接口
type Handler[A AggBase] interface {
	Handle(ctx context.Context, a A) error
}

type HandlerFunc[A AggBase] func(context.Context, A) error

func (f HandlerFunc[A]) Handle(ctx context.Context, a A) error {
	return f(ctx, a)
}

func NewHandler[A AggBase](hf func(context.Context, A) error) Handler[A] {
	return HandlerFunc[A](hf)
}

func Handle[A AggBase](ctx context.Context, h Handler[A], a A) (A, error) {
	if internal.InterfaceValNil(a) {
		return a, pkgerr.New("aggregate is nil")
	}
	slog.Debug("handle-begin", slog.Bool("isNew", a.IsNew()), slog.Group("agg", "id", a.ID(), "version", a.Version(), "createdAt", a.CreatedAt(), "updatedAt", a.UpdatedAt()))
	// 是否是新创建的聚合
	if a.IsNew() {
		if err := h.Handle(ctx, a); err != nil {
			return a, err
		}
	} else {
		cloned := clone.Clone(a)
		if err := h.Handle(ctx, a); err != nil {
			return a, err
		}
		var equals bool
		a.tempCleanEvents(func(events Events) {
			// 事件个数小于等于1的时候比较聚合的内容是否发生了变化
			if len(events) <= 1 {
				equals = reflect.DeepEqual(a, cloned)
			}
		})
		if equals {
			return a, nil
		} else {
			a.setUpdatedAt()
		}
	}

	// 验证
	if v, ok := AggBase(a).(Validator); ok {
		if err := v.Validate(); err != nil {
			return a, err
		}
	}

	// dry run
	if v, ok := h.(DryRunner); ok && v.DryRun() {
		return a, nil
	}
	// 补充事件属性
	a.completeEvents(a)
	// 更新版本
	if !a.changed() {
		a.incrVersion()
	}
	return a, nil
}

// BatchHandler 批量处理接口
type BatchHandler[A AggBase] interface {
	BatchHandle(ctx context.Context) ([]BatchEntry[A], error)
}

type BatchHandlerFunc[A AggBase] func(ctx context.Context) ([]BatchEntry[A], error)

func (f BatchHandlerFunc[A]) BatchHandle(ctx context.Context) ([]BatchEntry[A], error) {
	return f(ctx)
}

func NewBatchHandler[A AggBase](f func(ctx context.Context) ([]BatchEntry[A], error)) BatchHandler[A] {
	return BatchHandlerFunc[A](f)
}

// BatchEntry 批量命令返回条目
type BatchEntry[A AggBase] interface {
	Handler() Handler[A]
	ActionTarget() ActionTarget
}

// BatchEntry 批量命令返回条目
type batchEntry[A AggBase] struct {
	handler      Handler[A]
	actionTarget ActionTarget
}

func NewBatchEntry[A AggBase](handler Handler[A], target ActionTarget) *batchEntry[A] {
	return &batchEntry[A]{handler: handler, actionTarget: target}
}

func NewBatchEntryByFunc[A AggBase](hf func(ctx context.Context, a A) error, target ActionTarget) *batchEntry[A] {
	return NewBatchEntry[A](HandlerFunc[A](hf), target)
}

func (b batchEntry[A]) Handler() Handler[A] {
	return b.handler
}

func (b batchEntry[A]) ActionTarget() ActionTarget {
	return b.actionTarget
}

func HandleBatch[A AggBase](ctx context.Context, h BatchHandler[A], iterate func(context.Context, BatchEntry[A]) (A, error)) (as []A, err error) {
	entries, err := h.BatchHandle(ctx)
	if err != nil || len(entries) == 0 {
		return nil, err
	}
	if v, ok := h.(DryRunner); ok && v.DryRun() {
		return nil, nil
	}
	return internal.MapError(entries, func(i int, entry BatchEntry[A]) (A, error) {
		return iterate(ctx, entry)
	})
}
