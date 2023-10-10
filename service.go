package dgo

import (
	"context"

	"github.com/ekuu/dgo/internal"
	"github.com/ekuu/dgo/pb"
	pkgerr "github.com/pkg/errors"
)

type ActionTarget interface {
	isActionTarget()
}

type Service[A AggBase] interface {
	Get(ctx context.Context, id ID) (A, error)
	List(ctx context.Context, ids ...ID) ([]A, error)
	Create(ctx context.Context, h Handler[A]) (A, error)
	Delete(ctx context.Context, h Handler[A], t ActionTarget) error
	Update(ctx context.Context, h Handler[A], t ActionTarget) (A, error)
	Save(ctx context.Context, h Handler[A], t ActionTarget) (A, error)
	Batch(ctx context.Context, entries []*BatchEntry[A]) ([]A, error)
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

// Create 创建聚合
func (s *service[A]) Create(ctx context.Context, h Handler[A]) (a A, err error) {
	a, err = s.create(ctx, h)
	if err != nil {
		return a, err
	}
	return s.executeSaveOne(ctx, a)
}

func (s *service[A]) create(ctx context.Context, h Handler[A]) (a A, err error) {
	return s.create1(ctx, h, s.getAgg(h))
}

func (s *service[A]) create1(ctx context.Context, h Handler[A], a A) (A, error) {
	// 处理Handler
	a, err := handle(ctx, h, a)
	if err != nil {
		if v, ok := err.(*ErrDuplicate[A]); ok {
			return v.Aggregate(), nil
		}
		return a, err
	}

	// 生成一个id
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
func (s *service[A]) Delete(ctx context.Context, h Handler[A], t ActionTarget) error {
	a, err := s.delete(ctx, h, t)
	if err != nil {
		return err
	}

	return s.executeOne(ctx, a, func(ctx context.Context, r Repo[A]) error {
		return r.Delete(ctx, a)
	})
}

func (s *service[A]) delete(ctx context.Context, h Handler[A], t ActionTarget) (a A, err error) {
	a, err = s.getAggFromTarget(ctx, t)
	if err = IgnoreNotFound(err); err != nil {
		return a, err
	}
	return handle(ctx, h, a)
}

// Update 更新聚合
func (s *service[A]) Update(ctx context.Context, h Handler[A], t ActionTarget) (a A, err error) {
	a, err = s.update(ctx, h, t)
	if err != nil {
		return a, err
	}
	return s.executeSaveOne(ctx, a)
}

func (s *service[A]) update(ctx context.Context, h Handler[A], t ActionTarget) (A, error) {
	a, err := s.getAggFromTarget(ctx, t)
	if err != nil {
		return a, err
	}
	return handle(ctx, newUpdateHandler(h), a)
}

// Save 保存聚合
func (s *service[A]) Save(ctx context.Context, h Handler[A], t ActionTarget) (a A, err error) {
	a, err = s.save(ctx, h, t)
	if err != nil {
		return a, err
	}
	return s.executeSaveOne(ctx, a)
}

func (s *service[A]) save(ctx context.Context, h Handler[A], t ActionTarget) (a A, err error) {
	a, err = s.getAggFromTarget(ctx, t)
	if err = IgnoreIDNil(IgnoreNotFound(err)); err != nil {
		return a, err
	}
	if internal.InterfaceValNil(a) {
		a = s.getAgg(h)
	}
	if a.IsNew() {
		if v, ok := t.(ID); ok && a.ID().IsEmpty() {
			a.setID(v)
		}
		return s.create1(ctx, h, a)
	} else {
		return s.update(ctx, h, a)
	}
}

func (s *service[A]) Batch(ctx context.Context, entries []*BatchEntry[A]) (as []A, err error) {

	m := map[ActionType]func(context.Context, *BatchEntry[A]) (A, error){
		ActionCreate: func(ctx context.Context, e *BatchEntry[A]) (A, error) {
			return s.create(ctx, e.Handler)
		},
		ActionSave: func(ctx context.Context, e *BatchEntry[A]) (A, error) {
			return s.save(ctx, e.Handler, e.ActionTarget)
		},
		ActionUpdate: func(ctx context.Context, e *BatchEntry[A]) (A, error) {
			return s.update(ctx, e.Handler, e.ActionTarget)
		},
		ActionDelete: func(ctx context.Context, e *BatchEntry[A]) (A, error) {
			return s.delete(ctx, e.Handler, e.ActionTarget)
		},
	}
	// 检查ActionType
	for _, v := range entries {
		if _, ok := m[v.ActionType]; !ok {
			return nil, pkgerr.New("action type error")
		}
	}

	var es Events
	var fns []func(context.Context, Repo[A]) error
	for _, e := range entries {
		// 根据ActionType执行对应的Handler
		a, err := m[e.ActionType](ctx, e)
		if err != nil {
			return nil, err
		}
		as = append(as, a)

		// 聚合内容是否发生了变化
		if !a.changed() {
			continue
		}

		// 合并事件
		es = append(es, a.getEvents()...)

		if e.ActionType == ActionDelete {
			fns = append(fns, func(ctx context.Context, r Repo[A]) error {
				return r.Delete(ctx, a)
			})
		} else {
			fns = append(fns, func(ctx context.Context, r Repo[A]) error {
				return r.Save(ctx, a)
			})
		}
	}
	if len(fns) == 0 {
		return as, nil
	}

	// execute in transaction
	return as, s.transaction(ctx, es, func(ctx context.Context, r Repo[A]) error {
		for _, fn := range fns {
			if err = fn(ctx, r); err != nil {
				return err
			}
		}
		return nil
	})
}

type ActionType int

const (
	ActionCreate ActionType = iota + 1
	ActionUpdate
	ActionSave
	ActionDelete
)

// BatchEntry 批量命令返回条目
type BatchEntry[A AggBase] struct {
	Handler      Handler[A]
	ActionType   ActionType
	ActionTarget ActionTarget
}

func NewBatchEntry[A AggBase](handler Handler[A], actionType ActionType, actionTarget ActionTarget) *BatchEntry[A] {
	return &BatchEntry[A]{Handler: handler, ActionType: actionType, ActionTarget: actionTarget}
}

func NewBatchEntryByFunc[A AggBase](hf func(context.Context, A) error, actionType ActionType, actionTarget ActionTarget) *BatchEntry[A] {
	return &BatchEntry[A]{Handler: HandlerFunc[A](hf), ActionType: actionType, ActionTarget: actionTarget}
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
func (s *service[A]) getAgg(h Handler[A]) A {
	if v, ok := h.(AggConstructor[A]); ok {
		return v.NewAggregate()
	}
	return s.newAggregate()
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

//
//type BatchOptions struct {
//	transaction bool
//	parallel    bool // todo 批量并行
//	//continueWhenError bool
//}
//
//func WithBatchTransaction(transaction bool) func(o *BatchOptions) {
//	return func(o *BatchOptions) {
//		o.transaction = transaction
//	}
//}
//
//func WithBatchTransactionOff() func(o *BatchOptions) {
//	return func(o *BatchOptions) {
//		o.transaction = false
//	}
//}

//	func WithBatchContinueWhenError(continueWhenError bool) func(o *BatchOptions) {
//		return func(o *BatchOptions) {
//			o.continueWhenError = continueWhenError
//		}
//	}
//
//	func WithBatchReturnWhenError() func(o *BatchOptions) {
//		return func(o *BatchOptions) {
//			o.continueWhenError = false
//		}
//	}
