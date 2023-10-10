package dgo

import (
	"context"

	itrace "github.com/ekuu/dgo/internal/trace"
	"github.com/ekuu/dgo/pb"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var TraceSlog = itrace.SlogHandler

// repoTracer 为仓储逻辑添加trace支持
type repoTracer[A AggBase] struct {
	delegate Repo[A]
}

func traceRepo[A AggBase](r Repo[A]) Repo[A] {
	return &repoTracer[A]{delegate: r}
}

func (t *repoTracer[A]) Get(ctx context.Context, id ID) (A, error) {
	ctx, span := itrace.Start(ctx, "repo-get")
	defer span.End()
	a, err := t.delegate.Get(ctx, id)
	if err != nil {
		itrace.RecordError(span, IgnoreNotFound(err))
		return a, err
	}
	return a, nil
}

func (t *repoTracer[A]) List(ctx context.Context, ids ...ID) ([]A, error) {
	return itrace.Template2(ctx, "repo-list", func(ctx context.Context) ([]A, error) {
		strIDs := lo.Map(ids, func(item ID, index int) string {
			return item.String()
		})
		trace.SpanFromContext(ctx).SetAttributes(attribute.StringSlice("agg.ids", strIDs))
		return t.delegate.List(ctx, ids...)
	})
}

func (t *repoTracer[A]) Save(ctx context.Context, a A) error {
	return itrace.Template(ctx, "repo-save", func(ctx context.Context) error {
		return t.delegate.Save(ctx, a)
	})
}

func (t *repoTracer[A]) Delete(ctx context.Context, a A) error {
	return itrace.Template(ctx, "repo-delete", func(ctx context.Context) error {
		return t.delegate.Delete(ctx, a)
	})
}

func (t *repoTracer[A]) SaveEvents(ctx context.Context, events Events) error {
	return itrace.Template(ctx, "repo-events", func(ctx context.Context) error {
		return t.delegate.SaveEvents(ctx, events)
	})
}

func (t *repoTracer[A]) Transaction(ctx context.Context, fn func(ctx context.Context, r Repo[A]) error) error {
	return itrace.Template(ctx, "repo-trans", func(ctx context.Context) error {
		return t.delegate.Transaction(ctx, fn)
	})
}

// serviceTracer 为服务添加trace支持
type serviceTracer[A AggBase] struct {
	delegate Service[A]
}

func traceService[A AggBase](s Service[A]) Service[A] {
	return &serviceTracer[A]{delegate: s}
}

func (t *serviceTracer[A]) Get(ctx context.Context, id ID) (A, error) {
	return t.delegate.Get(ctx, id)
}

func (t *serviceTracer[A]) List(ctx context.Context, ids ...ID) ([]A, error) {
	return itrace.Template2(ctx, "service-list", func(ctx context.Context) ([]A, error) {
		return t.delegate.List(ctx, ids...)
	})

}

func (t *serviceTracer[A]) Create(ctx context.Context, h Handler[A]) (a A, err error) {
	return itrace.Template2(ctx, "service-create", func(ctx context.Context) (A, error) {
		return t.delegate.Create(ctx, h)
	})
}

func (t *serviceTracer[A]) Delete(ctx context.Context, h Handler[A], target ActionTarget) (err error) {
	return itrace.Template(ctx, "service-delete", func(ctx context.Context) error {
		return t.delegate.Delete(ctx, h, target)
	})
}

func (t *serviceTracer[A]) Update(ctx context.Context, h Handler[A], target ActionTarget) (A, error) {
	return itrace.Template2(ctx, "service-update", func(ctx context.Context) (A, error) {
		return t.delegate.Update(ctx, h, target)
	})
}

func (t *serviceTracer[A]) Save(ctx context.Context, h Handler[A], target ActionTarget) (a A, err error) {
	return itrace.Template2(ctx, "service-save", func(ctx context.Context) (A, error) {
		return t.delegate.Save(ctx, h, target)
	})
}

func (t *serviceTracer[A]) Batch(ctx context.Context, entries []*BatchEntry[A]) ([]A, error) {
	return itrace.Template2(ctx, "service-batch", func(ctx context.Context) ([]A, error) {
		return t.delegate.Batch(ctx, entries)
	})
}

// traceIDGenerator 为id生成添加trace支持
func traceIDGenerator(g IDGenerator) IDGenerator {
	return IDGenFunc(func(ctx context.Context) (ID, error) {
		return itrace.Template2(ctx, "id-generator", func(ctx context.Context) (ID, error) {
			id, err := g.GenID(ctx)
			if err == nil {
				trace.SpanFromContext(ctx).SetAttributes(attribute.String("agg.id", id.String()))
			}
			return id, err
		})
	})
}

// traceBus 为Bus添加trace支持
func traceBus(bus Bus) Bus {
	if nb, ok := bus.(NormalBus); ok {
		return traceNormalBus(nb)
	} else if tb, ok := bus.(TransactionBus); ok {
		return traceTransactionBus(tb)
	} else {
		return nil
	}
}

// normalBusTracer 为NormalBus添加trace支持
type normalBusTracer struct {
	delegate NormalBus
}

func traceNormalBus(b NormalBus) NormalBus {
	return &normalBusTracer{delegate: b}
}

func (t *normalBusTracer) Close() error {
	return t.delegate.Close()
}

func (t *normalBusTracer) Subscribe(ctx context.Context, rules ...*SubscribeRule) error {
	return itrace.Template(ctx, "bus-subscribe", func(ctx context.Context) error {
		return t.delegate.Subscribe(ctx, rules...)
	})
}

func (t *normalBusTracer) Publish(ctx context.Context, requests ...*pb.PublishRequest) error {
	return itrace.Template(ctx, "bus-normal-publish", func(ctx context.Context) error {
		return t.delegate.Publish(ctx, requests...)
	})
}

// TransactionBusTracer 为TransactionBus添加trace支持
type TransactionBusTracer struct {
	delegate TransactionBus
}

func traceTransactionBus(b TransactionBus) TransactionBus {
	return &TransactionBusTracer{delegate: b}
}

func (t *TransactionBusTracer) Close() error {
	return t.delegate.Close()
}

func (t *TransactionBusTracer) Subscribe(ctx context.Context, rules ...*SubscribeRule) error {
	return itrace.Template(ctx, "bus-subscribe", func(ctx context.Context) error {
		return t.delegate.Subscribe(ctx, rules...)
	})
}

func (t *TransactionBusTracer) TransactionPublish(ctx context.Context, repoTransaction func(ctx context.Context) error, requests ...*pb.PublishRequest) error {
	return itrace.Template(ctx, "bus-transaction-publish", func(ctx context.Context) error {
		return t.delegate.TransactionPublish(ctx, repoTransaction, requests...)
	})
}

// traceExecuteCallback 为事务仓储回调函数添加trace支持
func traceExecuteCallback[A AggBase](fn func(ctx context.Context, r Repo[A]) error) func(ctx context.Context, r Repo[A]) error {
	return func(ctx context.Context, r Repo[A]) error {
		return itrace.Template(ctx, "execute-callback", func(ctx context.Context) error {
			return fn(ctx, r)
		})
	}
}
