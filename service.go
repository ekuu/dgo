package dgo

import (
	"context"

	"github.com/ekuu/dgo/internal"
	"github.com/ekuu/dgo/pb"
	pkgerr "github.com/pkg/errors"
	"github.com/samber/lo"
)

type ActionTarget interface {
	isActionTarget()
}

type Service[A AggBase] interface {
	Get(ctx context.Context, id ID) (A, error)
	List(ctx context.Context, ids ...ID) ([]A, error)
	Create(ctx context.Context, h Handler[A]) (a A, err error)
	Delete(ctx context.Context, h Handler[A], t ActionTarget) (err error)
	Update(ctx context.Context, h Handler[A], t ActionTarget) (A, error)
	Save(ctx context.Context, h Handler[A], t ActionTarget) (a A, err error)
	BatchCreate(ctx context.Context, h BatchHandler[A]) ([]A, error)
	BatchUpdate(ctx context.Context, h BatchHandler[A]) error
	BatchSave(ctx context.Context, h BatchHandler[A]) ([]A, error)
	BatchDelete(ctx context.Context, h BatchHandler[A]) error
}

func NewService[A AggBase](repo Repo[A], newAggregate func() A, opts ...ServiceOption[A]) Service[A] {
	return traceService[A](newService(repo, newAggregate, opts...))
}

// Service 带有事务的仓储接口包装，用于处理通用逻辑
//
//go:generate gogen option -n service -r repo,newAggregate --with-init --lowercase
type service[A AggBase] struct {
	repo         Repo[A]
	bus          Bus
	idGenerator  IDGenerator
	newAggregate func() A
}

func WithServiceIdGenFunc[A AggBase](f func(ctx context.Context) (ID, error)) ServiceOption[A] {
	return serviceOptionFunc[A](func(c *service[A]) {
		c.idGenerator = IDGenFunc(f)
	})
}

func (s *service[A]) init() {
	if s.idGenerator == nil {
		s.idGenerator = IDGenFunc(GenNoHyphenUUID)
	}
	s.bus = traceBus(s.bus)
	s.repo = traceRepo(s.repo)
	s.idGenerator = traceIDGenerator(s.idGenerator)
}

// Get 查找聚合
// 如未找到,返回"ErrNotFound"错误
func (s *service[A]) Get(ctx context.Context, id ID) (A, error) {
	return s.repo.Get(ctx, id)
}

// List 查找多个聚合
func (s *service[A]) List(ctx context.Context, ids ...ID) ([]A, error) {
	return s.repo.List(ctx, ids...)
}

// Create 处理创建命令
func (s *service[A]) Create(ctx context.Context, h Handler[A]) (a A, err error) {
	// create
	a, err = s.create(ctx, h, s.getAggConstruct(h)())
	if err != nil {
		return a, err
	}
	// execute
	return s.executeSaveOne(ctx, a)
}

func (s *service[A]) create(ctx context.Context, h Handler[A], a A) (A, error) {
	a, err := Handle(ctx, h, a)
	if err != nil {
		return a, err
	}
	if !a.IsNew() {
		return a, nil
	}
	if a.ID().IsEmpty() {
		id, err := s.idGenerator.GenID(ctx)
		if err != nil {
			return a, err
		}
		a.setID(id)
	}
	return a, nil
}

// Delete 处理删除命令
func (s *service[A]) Delete(ctx context.Context, h Handler[A], t ActionTarget) (err error) {
	a, err := s.getAggFromTarget(ctx, t)
	if err = IgnoreNotFound(err); err != nil {
		return err
	}

	a, err = Handle(ctx, h, a)
	if err != nil {
		return err
	}

	return s.executeOne(ctx, a, func(ctx context.Context, r Repo[A]) error {
		return r.Delete(ctx, a)
	})
}

// Update 处理更新命令
func (s *service[A]) Update(ctx context.Context, h Handler[A], t ActionTarget) (A, error) {
	a, err := s.getAggFromTarget(ctx, t)
	if err != nil {
		return a, err
	}

	a, err = Handle(ctx, h, a)
	if err != nil {
		return a, err
	}
	return s.executeSaveOne(ctx, a)
}

// Save 聚合存在则更新，不存在则创建
func (s *service[A]) Save(ctx context.Context, h Handler[A], t ActionTarget) (a A, err error) {
	a, err = s.save(ctx, h, s.getAggConstruct(h), t)
	if err != nil {
		return a, err
	}
	return s.executeSaveOne(ctx, a)
}

func (s *service[A]) save(ctx context.Context, h Handler[A], newAgg func() A, t ActionTarget) (A, error) {
	a, err := s.getAggFromTarget(ctx, t)
	if err = IgnoreIDNil(IgnoreNotFound(err)); err != nil {
		return a, err
	}
	if internal.InterfaceValNil(a) {
		a = newAgg()
	}
	if a.IsNew() {
		if id, ok := t.(ID); ok && a.ID().IsEmpty() {
			a.setID(id)
		}
		return s.create(ctx, h, a)
	} else {
		return Handle(ctx, h, a)
	}
}

func (s *service[A]) handleBatchSave(ctx context.Context, h BatchHandler[A], iterate func(context.Context, BatchEntry[A]) (A, error)) ([]A, error) {
	as, err := HandleBatch[A](ctx, h, iterate)
	if err != nil || len(as) == 0 {
		return nil, err
	}
	return s.transactionSaveMany(ctx, as)
}

// BatchCreate 创建多个聚合
// 创建在数据库中不存在的聚合，如果已存在于数据库中（如聚合中的某些字段对应为数据库唯一索引）则返回命中的数据
func (s *service[A]) BatchCreate(ctx context.Context, h BatchHandler[A]) ([]A, error) {
	return s.handleBatchSave(
		ctx,
		h,
		func(ctx context.Context, e BatchEntry[A]) (a A, err error) {
			if t := e.ActionTarget(); t != nil {
				var ok bool
				if a, ok = t.(A); !ok {
					a = s.getAggConstruct(h)()
					a.setID(t.(ID))
				}
			} else {
				a = s.getAggConstruct(h)()
			}
			return s.create(ctx, e.Handler(), a)
		},
	)
}

// BatchUpdate 保存多个聚合
func (s *service[A]) BatchUpdate(ctx context.Context, h BatchHandler[A]) error {
	_, err := s.handleBatchSave(
		ctx,
		h,
		func(ctx context.Context, e BatchEntry[A]) (A, error) {
			a, err := s.getAggFromTarget(ctx, e.ActionTarget())
			if err != nil {
				return a, err
			}
			return Handle(ctx, e.Handler(), a)
		},
	)
	return err
}

// BatchSave 保存多个聚合
func (s *service[A]) BatchSave(ctx context.Context, h BatchHandler[A]) ([]A, error) {
	return s.handleBatchSave(
		ctx,
		h,
		func(ctx context.Context, e BatchEntry[A]) (A, error) {
			return s.save(ctx, e.Handler(), s.getAggConstruct(h), e.ActionTarget())
		},
	)
}

// BatchDelete 删除多个聚合
func (s *service[A]) BatchDelete(ctx context.Context, h BatchHandler[A]) error {
	iterate := func(ctx context.Context, e BatchEntry[A]) (A, error) {
		a, err := s.getAggFromTarget(ctx, e.ActionTarget())
		if err != nil {
			return a, IgnoreNotFound(err)
		}
		return Handle(ctx, e.Handler(), a)
	}

	es, err := HandleBatch[A](ctx, h, iterate)
	if err != nil || len(es) == 0 {
		return err
	}

	_, err = s.transactionMany(ctx, es, func(ctx context.Context, r Repo[A], as []A) error {
		for _, a := range as {
			if err := r.Delete(ctx, a); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (s *service[A]) getAggFromTarget(ctx context.Context, t ActionTarget) (a A, err error) {
	if t == nil {
		return
	}
	var ok bool
	if a, ok = t.(A); !ok {
		if id := t.(ID); id.IsEmpty() {
			return a, pkgerr.WithStack(ErrIDNil)
		} else {
			return s.Get(ctx, id)
		}
	}
	return
}

func (s *service[A]) getAggConstruct(h any) func() A {
	if v, ok := h.(AggConstructor[A]); ok {
		return v.NewAggregate
	}
	return s.newAggregate
}

// transaction 事务
func (s *service[A]) transaction(ctx context.Context, es Events, fn func(ctx context.Context, r Repo[A]) error) error {
	// transaction
	transaction := func(ctx context.Context) error {
		return s.repo.Transaction(ctx, func(ctx context.Context, r Repo[A]) error {
			if err := r.SaveEvents(ctx, es); err != nil {
				return err
			}
			return traceExecuteCallback(fn)(ctx, s.repo)
		})
	}

	// build publish requests
	requests, err := internal.MapError(es, func(i int, e *Event) (*pb.PublishRequest, error) {
		return e.PublishRequest()
	})
	if err != nil {
		return err
	}
	if s.bus != nil {
		// assert EventBus
		if normalBus, ok := s.bus.(NormalBus); ok {
			if err := transaction(ctx); err != nil {
				return err
			}
			return normalBus.Publish(ctx, requests...)
		} else if transactionBus, ok := s.bus.(TransactionBus); ok {
			return transactionBus.TransactionPublish(ctx, transaction, requests...)
		} else {
			return pkgerr.New("Event bus assert fail")
		}
	} else {
		return transaction(ctx)
	}
}

// executeOne 执行聚合仓储操作,内部判断是否需要开启事务
func (s *service[A]) executeOne(ctx context.Context, a A, fn func(ctx context.Context, r Repo[A]) error) error {
	if s.needTransaction(a) {
		return s.transaction(ctx, a.getEvents(), fn)
	}
	return traceExecuteCallback(fn)(ctx, s.repo)
}

// executeSaveOne 执行聚合的保存动作
func (s *service[A]) executeSaveOne(ctx context.Context, a A) (A, error) {
	if !a.changed() {
		return a, nil
	}
	return a, s.executeOne(ctx, a, func(ctx context.Context, r Repo[A]) error {
		return r.Save(ctx, a)
	})
}

// needTransaction 判断是否需要开启事务
func (s *service[A]) needTransaction(a A) bool {
	// 有事件产生
	if len(a.getEvents()) > 0 {
		return true
	}
	// 实现了MultiDocuments接口
	if _, ok := AggBase(a).(MultiDocuments); ok {
		return true
	}
	return false
}

// transactionMany 执行多个事件的事务,忽略无payload的事件
func (s *service[A]) transactionMany(ctx context.Context, as []A, fn func(context.Context, Repo[A], []A) error) ([]A, error) {
	changedAggregates := lo.Filter(as, func(a A, index int) bool {
		return a.changed()
	})
	if len(as) == 0 {
		return as, nil
	}
	var es Events
	for _, a := range changedAggregates {
		es = append(es, a.getEvents()...)
	}
	return as, s.transaction(ctx, es, func(ctx context.Context, r Repo[A]) error {
		return fn(ctx, r, as)
	})
}

// transactionSaveMany 执行多个事件及聚合保存动作的事务
func (s *service[A]) transactionSaveMany(ctx context.Context, as []A) ([]A, error) {
	return s.transactionMany(ctx, as, func(ctx context.Context, r Repo[A], as []A) error {
		for _, a := range as {
			if err := r.Save(ctx, a); err != nil {
				return err
			}
		}
		return nil
	})
}
