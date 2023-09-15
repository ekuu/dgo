package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ekuu/dgo"

	"github.com/ekuu/dgo/pb"
)

func TestEventBus(t *testing.T) {
	eb := NewMemory(10)
	ss := []*dgo.SubscribeRule{
		dgo.NewSubscribeRule("topic1", func(ctx context.Context, e *pb.Event) error {
			fmt.Printf("SubscribeRule 1 received event: %+v\n", e)
			return nil
		}),
		dgo.NewSubscribeRule("topic2", func(ctx context.Context, e *pb.Event) error {
			fmt.Printf("SubscribeRule 2 received event: %+v\n", e)
			return nil
		}),
		dgo.NewSubscribeRule("topic2", func(ctx context.Context, e *pb.Event) error {
			fmt.Printf("SubscribeRule 3 received event: %+v\n", e)
			return nil
		}),
	}
	// 订阅者3订阅topic2
	if err := eb.Subscribe(context.Background(), ss...); err != nil {
		t.Fatalf("Failed to subscribe %v", err)
	}
	// 发布一个事件到topic1
	r1 := &pb.PublishRequest{Topic: "topic1", Event: &pb.Event{Name: "Event 1"}}
	r11 := &pb.PublishRequest{Topic: "topic1", Event: &pb.Event{Name: "Event 11"}}
	if err := eb.Publish(context.Background(), r1, r11); err != nil {
		t.Fatalf("Failed to publish event to topic1: %v", err)
	}

	// 发布一个事件到topic2
	r2 := &pb.PublishRequest{Topic: "topic2", Event: &pb.Event{Name: "Event 2"}}
	r22 := &pb.PublishRequest{Topic: "topic2", Event: &pb.Event{Name: "Event 22"}}
	if err := eb.Publish(context.Background(), r2, r22); err != nil {
		t.Fatalf("Failed to publish event to topic2: %v", err)
	}

	// 等待事件处理完成
	time.Sleep(time.Millisecond * 100)
}
