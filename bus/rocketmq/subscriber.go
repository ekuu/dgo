package rocketmq

import (
	"context"

	"github.com/ekuu/dgo/pb"
	"google.golang.org/protobuf/encoding/protojson"

	mq "github.com/apache/rocketmq-client-go/v2"
	mqc "github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/ekuu/dgo"
	pkgerr "github.com/pkg/errors"
)

type subscriber struct {
	consumer mq.PushConsumer
}

func NewSubscriber(opts ...mqc.Option) (*subscriber, error) {
	c, err := mq.NewPushConsumer(opts...)
	if err != nil {
		return nil, pkgerr.Wrap(err, "failed to create consumer")
	}
	return &subscriber{consumer: c}, nil
}

// Subscribe 函数用于订阅消息
func (c *subscriber) Subscribe(ctx context.Context, rules ...*dgo.SubscribeRule) error {
	for _, rule := range rules {
		handle := rule.Handle
		if err := c.consumer.Subscribe(rule.Topic, mqc.MessageSelector{}, func(ctx context.Context, exts ...*primitive.MessageExt) (mqc.ConsumeResult, error) {
			for _, ext := range exts {
				e := new(pb.Event)
				if err := protojson.Unmarshal(ext.Body, e); err != nil {
					return mqc.Rollback, pkgerr.New("fail to unmarshal event")
				}
				if err := handle(ctx, e); err != nil {
					return mqc.Rollback, pkgerr.New("fail to handle event")
				}
			}
			return mqc.ConsumeSuccess, nil
		}); err != nil {
			return pkgerr.Wrapf(err, "failed to subscribe to topic: %s", rule.Topic)
		}
	}
	return c.consumer.Start()
}

func (c *subscriber) Close() error {
	return c.consumer.Shutdown()
}
