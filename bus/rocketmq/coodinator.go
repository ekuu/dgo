package rocketmq

import (
	"context"
	"sync"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/google/uuid"
	pkgerr "github.com/pkg/errors"
	"github.com/samber/lo"
)

// 协调RocketMQ消息的发送与仓储逻辑的执行
type coordinator struct {
	id            uuid.UUID
	messages      []*primitive.Message
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        func()
	errOnce       sync.Once
	err           error
	repoTransDone chan interface{}
}

func newCoordinator(messages []*primitive.Message) *coordinator {
	ctx, cancel := context.WithCancel(context.Background())
	return &coordinator{
		id:            uuid.New(),
		messages:      messages,
		ctx:           ctx,
		cancel:        cancel,
		repoTransDone: make(chan interface{}),
	}
}

// ID 获取id
func (c *coordinator) ID() string {
	return c.id.String()
}

// GO 启动一个goroutine发送RocketMQ消息
func (c *coordinator) GO(f func() (*primitive.TransactionSendResult, error)) {
	c.wg.Add(1)
	// todo ho goroutine
	go func() {
		rs, err := f()
		if err == nil && rs.Status != primitive.SendOK {
			err = pkgerr.New(rs.String())
		}
		defer func() {
			if err != nil {
				c.wg.Done()
			}
		}()
		if err != nil {
			c.errOnce.Do(func() {
				c.err = err
				if c.cancel != nil {
					c.cancel()
				}
			})
		}
	}()
}

// WaitAllOK 等待所有goroutine执行完毕, 返回结果为是否全部消息都发送成功
func (c *coordinator) WaitAllOK() error {
	c.wg.Wait()
	if c.cancel != nil {
		c.cancel()
	}
	if c.err != nil {
		return c.err
	}
	count := lo.CountBy(c.messages, func(m *primitive.Message) bool { return m.GetProperty("SendOK") == "1" })
	if count != len(c.messages) {
		return pkgerr.Errorf("send transaction events fail, expect:%d ,ok:%d", len(c.messages), count)
	}
	return nil
}

// SendOK 为消息标记发送成功，然后等待仓储事务执行
func (c *coordinator) SendOK(m *primitive.Message) error {
	m.WithProperty("SendOK", "1")
	c.wg.Done()
	<-c.repoTransDone
	return c.err
}

// ExecRepoTrans 执行仓储事务
func (c *coordinator) ExecRepoTrans(ctx context.Context, repoTrans func(ctx context.Context) error) error {
	c.err = repoTrans(ctx)
	close(c.repoTransDone)
	return c.err
}
