package dgo

import (
	"context"
	"log/slog"
	"reflect"

	"github.com/ekuu/dgo/internal"
	itrace "github.com/ekuu/dgo/internal/trace"
	clone "github.com/huandu/go-clone/generic"
	pkgerr "github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
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

// handle 执行Handler接口的Handle函数
func handle[A AggBase](ctx context.Context, h Handler[A], a A) (A, error) {
	if internal.InterfaceValNil(a) {
		return a, pkgerr.New("aggregate is nil")
	}

	ctx, span := itrace.Start(ctx, "handle")
	defer span.End()

	slog.InfoContext(ctx, "handle start", slog.Bool("isNew", a.IsNew()), slog.Group("agg", "id", a.ID(), "name", getAggName(a), "version", a.Version(), "createdAt", a.CreatedAt(), "updatedAt", a.UpdatedAt()))

	if err := h.Handle(ctx, a); err != nil {
		return a, err
	}

	// 验证
	if v, ok := AggBase(a).(Validator); ok {
		if err := v.Validate(); err != nil {
			slog.InfoContext(ctx, "aggregate validate", attribute.String("message", err.Error()))
			return a, err
		}
	}

	// dry run
	if v, ok := h.(DryRunner); ok && v.DryRun() {
		slog.DebugContext(ctx, "command dry run", slog.Bool("result", true))
		return a, nil
	}

	// 补充事件属性
	a.base().completeEvents(a)
	// 更新版本
	if !a.changed() {
		a.base().incrVersion()
	}
	return a, nil
}

// updateHandler 对更新操作前后内容进行对比
type updateHandler[A AggBase] struct {
	Handler[A]
}

func newUpdateHandler[A AggBase](h Handler[A]) Handler[A] {
	return &updateHandler[A]{Handler: h}
}

func (h updateHandler[A]) Handle(ctx context.Context, a A) error {
	cloned := clone.Clone(a)
	if err := h.Handler.Handle(ctx, a); err != nil {
		return err
	}

	diff := true
	var attrs []any

	a.base().tempCleanEvents(func(events Events) {
		// 事件个数小于等于1的时候比较聚合的内容是否发生了变化
		attrs = append(attrs, slog.Int("eventCount", len(events)))
		if len(events) <= 1 {
			diff = !reflect.DeepEqual(a, cloned)
		}
	})

	slog.DebugContext(ctx, "update aggregate", append(attrs, slog.Bool("changed", diff))...)

	if diff {
		a.base().setUpdatedAtNow()
	}

	return nil
}
