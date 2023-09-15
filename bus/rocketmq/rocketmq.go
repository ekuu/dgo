package rocketmq

import (
	mqc "github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	mqp "github.com/apache/rocketmq-client-go/v2/producer"
)

type Producer interface {
	Start() error
	Shutdown() error
}

// rocketMQ 结构体用于包装 RocketMQ 的生产者和消费者
//
//go:generate gogen option -n normalRocket -e common -s getMetadata --lowercase
type normalRocket struct {
	*subscriber
	*normalPublisher
}

// NewNormalRocketMQ 函数用于创建一个 rocketMQ 结构体实例
func NewNormalRocketMQ(producerOptions []mqp.Option, consumerOptions []mqc.Option, opts ...PubOption) (*normalRocket, error) {
	s, err := NewSubscriber(consumerOptions...)
	if err != nil {
		return nil, err
	}
	p, err := NewNormalPublisher(producerOptions, opts...)
	if err != nil {
		return nil, err
	}
	return &normalRocket{subscriber: s, normalPublisher: p}, nil
}

func (r *normalRocket) Close() error {
	// TODO log
	r.subscriber.Close()
	return r.normalPublisher.Close()
}

//go:generate gogen option -n transactionRocket -e common -s getMetadata --lowercase
type transactionRocket struct {
	*subscriber
	*transactionPublisher
}

func NewTransactionRocketMQ(
	producerOptions []mqp.Option,
	consumerOptions []mqc.Option,
	checkLocalTransaction func(ext *primitive.MessageExt) primitive.LocalTransactionState,
	opts ...PubOption,
) (*transactionRocket, error) {
	s, err := NewSubscriber(consumerOptions...)
	if err != nil {
		return nil, err
	}
	p, err := NewTransactionPublisher(producerOptions, checkLocalTransaction, opts...)
	if err != nil {
		return nil, err
	}
	return &transactionRocket{subscriber: s, transactionPublisher: p}, nil
}

func (r *transactionRocket) Close() error {
	// TODO log
	r.subscriber.Close()
	return r.transactionPublisher.Close()
}
