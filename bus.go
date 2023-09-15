package dgo

import (
	"context"
	"io"

	"github.com/ekuu/dgo/pb"
)

// Bus 事件总线
type Bus interface {
	Subscriber
}

// NormalBus 普通总线
type NormalBus interface {
	Bus
	NormalPublisher
}

// TransactionBus 事务总线
type TransactionBus interface {
	Bus
	TransactionPublisher
}

type NormalPublisher interface {
	io.Closer
	Publish(ctx context.Context, requests ...*pb.PublishRequest) error
}

type TransactionPublisher interface {
	io.Closer
	TransactionPublish(ctx context.Context, repoTransaction func(ctx context.Context) error, requests ...*pb.PublishRequest) error
}

type Subscriber interface {
	io.Closer
	Subscribe(ctx context.Context, rules ...*SubscribeRule) error
}

// SubscribeRule 订阅规则
type SubscribeRule struct {
	Topic  string
	Handle func(ctx context.Context, e *pb.Event) error
}

func NewSubscribeRule(topic string, handle func(ctx context.Context, e *pb.Event) error) *SubscribeRule {
	return &SubscribeRule{Topic: topic, Handle: handle}
}
