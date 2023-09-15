package rocketmq

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ekuu/dgo"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/ekuu/dgo/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewRocketMQ(t *testing.T) {
	producerOptions := []producer.Option{
		producer.WithNameServer([]string{"192.168.31.210:9876"}),
		producer.WithGroupName(fmt.Sprintf("%s-event-producer", "ddd-test")),
		producer.WithRetry(2),
	}
	consumerOptions := []consumer.Option{
		consumer.WithNameServer([]string{"192.168.31.210:9876"}),
		consumer.WithGroupName(fmt.Sprintf("%s-push-consumer", "ddd-test")),
	}
	// 创建 rocketMQ 实例
	mq, err := NewNormalRocketMQ(producerOptions, consumerOptions)
	if err != nil {
		t.Errorf("NewNormalRocketMQ failed: %v", err)
		return
	}
	topic1 := "TestTopic1"
	err = mq.Publish(context.Background(), &pb.PublishRequest{
		Topic: topic1,
		Event: &pb.Event{
			Name:      "test-name",
			AggId:     "test-agg-id",
			CreatedAt: timestamppb.New(time.Now()),
		},
	})
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}

	topic2 := "TestTopic2"
	err = mq.Publish(context.Background(), &pb.PublishRequest{
		Topic: topic2,
		Event: &pb.Event{
			Name:      "test-name2",
			AggId:     "test-agg-id2",
			CreatedAt: timestamppb.New(time.Now()),
		},
	})
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}

	ss := []*dgo.SubscribeRule{
		dgo.NewSubscribeRule(topic1, func(ctx context.Context, e *pb.Event) error {
			fmt.Println("--------------------------------------------------------")
			fmt.Printf("%s: %v\n", topic1, e)
			fmt.Println("--------------------------------------------------------")
			return nil
		}),
		dgo.NewSubscribeRule(topic2, func(ctx context.Context, e *pb.Event) error {
			fmt.Println("--------------------------------------------------------")
			fmt.Printf("%s: %v\n", topic2, e)
			fmt.Println("--------------------------------------------------------")
			return nil
		}),
	}
	if err = mq.Subscribe(context.Background(), ss...); err != nil {
		t.Error(err)
	}
	time.Sleep(1 * time.Second)
	// 关闭 rocketMQ 实例
	if err := mq.Close(); err != nil {
		t.Errorf("Failed to close rocketMQ: %v", err)
	}

}
