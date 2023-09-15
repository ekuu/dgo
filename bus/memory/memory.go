package memory

import (
	"context"
	"log"

	"github.com/ekuu/dgo"
	"github.com/ekuu/dgo/pb"
	"github.com/samber/lo"
)

type memory struct {
	bufferSize int
	handlers   map[string][]chan *pb.Event
}

func NewMemory(bufferSize int) *memory {
	return &memory{bufferSize: bufferSize, handlers: make(map[string][]chan *pb.Event)}
}

func (m *memory) Publish(ctx context.Context, requests ...*pb.PublishRequest) error {

	// 向所有handler发送事件
	for _, r := range requests {
		// 获取topic的所有handlers
		chans, ok := m.handlers[r.Topic]
		if !ok {
			// 如果没有handler则直接返回
			return nil
		}
		for _, ch := range chans {
			select {
			case ch <- r.Event:
			default:
				// 如果chan满了则丢弃事件
				log.Printf("memory: event dropped, topic=%s", r.Topic)
			}
		}
	}

	return nil
}

func (m *memory) Subscribe(ctx context.Context, subscribers ...*dgo.SubscribeRule) error {
	sm := lo.GroupBy(subscribers, func(subscriber *dgo.SubscribeRule) string {
		return subscriber.Topic
	})
	for topic, ss := range sm {
		for _, s := range ss {
			ch := make(chan *pb.Event, m.bufferSize)
			m.handlers[topic] = append(m.handlers[topic], ch)
			go func(subscriber *dgo.SubscribeRule) {
				for e := range ch {
					if err := subscriber.Handle(context.Background(), e); err != nil {
						// 处理事件的handler返回错误时，记录日志
						log.Printf("memory: event handling failed, topic=%s, err=%v", topic, err)
					}
				}
			}(s)
		}
	}
	return nil
}

func (m *memory) Close() error {
	// 关闭所有channel
	for topic, chans := range m.handlers {
		for _, ch := range chans {
			close(ch)
		}
		delete(m.handlers, topic)
	}

	return nil
}
