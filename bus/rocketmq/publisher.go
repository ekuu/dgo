package rocketmq

import (
	"context"
	"strings"
	"sync"

	"github.com/ekuu/dgo/pb"
	"google.golang.org/protobuf/encoding/protojson"

	mq "github.com/apache/rocketmq-client-go/v2"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	mqp "github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/ekuu/dgo/internal"
	pkgerr "github.com/pkg/errors"
)

// rocketMQ 结构体用于包装 RocketMQ 的生产者和消费者
type normalPublisher struct {
	*pubCommon[mq.Producer]
}

// NewNormalPublisher 函数用于创建一个 normalPublisher 结构体实例
func NewNormalPublisher(producerOptions []mqp.Option, opts ...PubOption) (*normalPublisher, error) {
	c, err := newPubCommon(
		producerOptions,
		func() mq.Producer { return nil },
		func(options []mqp.Option) (mq.Producer, error) {
			return mq.NewProducer(options...)
		},
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &normalPublisher{pubCommon: c}, nil
}

// Publish 函数用于发送消息
func (r *normalPublisher) Publish(ctx context.Context, es ...*pb.PublishRequest) error {
	messages, err := internal.MapError(es, func(i int, e *pb.PublishRequest) (*primitive.Message, error) {
		return r.newMessage(e)
	})
	if err != nil {
		return err
	}
	producer, err := r.getProducer()
	if err != nil {
		return pkgerr.Wrap(err, "rocketmq message send fail because producer failed to initialize")
	}
	if _, err := producer.SendSync(ctx, messages...); err != nil {
		r.resetProducer()
		return pkgerr.Wrap(err, "failed to send message")
	}
	return nil
}

type transactionPublisher struct {
	*pubCommon[mq.TransactionProducer]
	coordinators          sync.Map
	checkLocalTransaction func(ext *primitive.MessageExt) primitive.LocalTransactionState
}

func NewTransactionPublisher(
	producerOptions []mqp.Option,
	checkLocalTransaction func(ext *primitive.MessageExt) primitive.LocalTransactionState,
	opts ...PubOption,
) (*transactionPublisher, error) {
	publisher := &transactionPublisher{
		coordinators:          sync.Map{},
		checkLocalTransaction: checkLocalTransaction,
	}
	c, err := newPubCommon(
		producerOptions,
		func() mq.TransactionProducer { return nil },
		func(options []mqp.Option) (mq.TransactionProducer, error) {
			return mq.NewTransactionProducer(publisher, options...)
		},
		opts...,
	)
	if err != nil {
		return nil, err
	}
	publisher.pubCommon = c
	return publisher, nil
}

func (r *transactionPublisher) TransactionPublish(ctx context.Context, repoTrans func(ctx context.Context) error, requests ...*pb.PublishRequest) error {
	messages, err := internal.MapError(requests, func(i int, req *pb.PublishRequest) (*primitive.Message, error) {
		return r.newMessage(req)
	})
	if err != nil {
		return err
	}
	producer, err := r.getProducer()
	if err != nil {
		return pkgerr.Wrap(err, "rocketmq message send fail because producer failed to initialize")
	}
	// 协调RocketMQ与仓储逻辑，每发送一批消息均启用一个协调者
	c := newCoordinator(messages)
	r.coordinators.Store(c.ID(), c)
	defer r.coordinators.Delete(c.ID())

	// 发送消息
	for _, m := range c.messages {
		message := m
		m.WithProperty("CoordinatorID", c.ID())
		c.GO(func() (*primitive.TransactionSendResult, error) {
			rs, err := producer.SendMessageInTransaction(ctx, message)
			if err != nil {
				r.resetProducer()
				return rs, pkgerr.Wrap(err, "failed to send message")
			}
			return rs, nil
		})
	}

	// 等待所有消息在ExecuteLocalTransaction中执行SendOK，如有任意一个消息未发送至RocketMQ则不执行仓储逻辑
	if err = c.WaitAllOK(); err != nil {
		return err
	}
	return c.ExecRepoTrans(ctx, repoTrans)
}

func (r *transactionPublisher) ExecuteLocalTransaction(message *primitive.Message) primitive.LocalTransactionState {
	// 获取协调者信息
	rs, ok := r.coordinators.Load(message.GetProperty("CoordinatorID"))
	if !ok {
		return primitive.RollbackMessageState
	}
	c, ok := rs.(*coordinator)
	if !ok {
		return primitive.RollbackMessageState
	}
	// 标记消息已发送成功、等待仓储逻辑的执行结果
	if err := c.SendOK(message); err != nil {
		return primitive.RollbackMessageState
	} else {
		return primitive.CommitMessageState
	}
}

func (r *transactionPublisher) CheckLocalTransaction(ext *primitive.MessageExt) primitive.LocalTransactionState {
	return r.checkLocalTransaction(ext)
}

//go:generate gogen option -n pubConfig -p pub --lowercase --with-init
type pubConfig struct {
	getMetadata func(e *pb.Event) Metadata
}

func (p *pubConfig) init() {
	if p.getMetadata == nil {
		p.getMetadata = func(e *pb.Event) Metadata {
			return Metadata{
				MetadataRocketmqKey: e.Name,
				MetadataRocketmqTag: e.Name,
			}
		}
	}
}

type pubCommon[P Producer] struct {
	*pubConfig
	producer        P
	getNilProducer  func() P
	newProducer     func(options []mqp.Option) (P, error)
	lock            sync.RWMutex // 生产者锁，用于确保多个协程同时访问生产者时的数据一致性
	producerOptions []mqp.Option // 用于初始化生产者的选项
}

func newPubCommon[P Producer](
	producerOptions []mqp.Option,
	getNilProducer func() P,
	newProducer func(options []mqp.Option) (P, error),
	pubOptions ...PubOption,
) (*pubCommon[P], error) {
	return &pubCommon[P]{
		newProducer:     newProducer,
		getNilProducer:  getNilProducer,
		producerOptions: producerOptions,
		pubConfig:       newPub(pubOptions...),
	}, nil
}

func (c *pubCommon[P]) Close() error {
	p, err := c.getProducer()
	if err != nil {
		return err
	}
	return p.Shutdown()
}

// resetProducer 函数用于重置生产者，当生产者出错时会调用
func (c *pubCommon[P]) resetProducer() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.producer = c.getNilProducer()
}

// getProducer 函数用于获取生产者实例
func (c *pubCommon[P]) getProducer() (p P, err error) {
	if !internal.InterfaceValNil(c.producer) {
		return c.producer, nil
	}
	c.lock.Lock()
	defer c.lock.Unlock()

	if !internal.InterfaceValNil(c.producer) {
		return c.producer, nil
	}
	producer, err := c.newProducer(c.producerOptions)
	if err != nil {
		return p, pkgerr.Wrap(err, "failed to create producer")
	}
	err = producer.Start()
	if err != nil {
		_ = producer.Shutdown()
		return p, err
	}
	c.producer = producer
	return c.producer, nil
}

func (c *pubCommon[P]) newMessage(req *pb.PublishRequest) (*primitive.Message, error) {
	body, err := protojson.Marshal(req.Event)
	if err != nil {
		return nil, pkgerr.Wrap(err, "failed to marshal event")
	}
	m := primitive.NewMessage(req.Topic, body)
	for k, v := range c.getMetadata(req.Event) {
		switch strings.ToLower(k) {
		case MetadataRocketmqTag:
			m.WithTag(v)
		case MetadataRocketmqKey:
			m.WithKeys(strings.Split(v, ","))
		case MetadataRocketmqShardingKey:
			m.WithShardingKey(v)
		}
	}
	return m, nil
}
