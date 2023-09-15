package order

import (
	"context"

	"github.com/ekuu/dgo/internal/examples/infra/types"
	"github.com/pkg/errors"
)

type CreateCmd struct {
	Items Items
	Off   uint8
}

func (c *CreateCmd) Handle(ctx context.Context, o *Order) error {
	if c.Off > 30 {
		// 这里可以使用自己定义的错误返回接口
		return errors.New("优惠力度不得大于30%")
	}
	o.items = c.Items
	o.totalPrice = c.Items.TotalPrice()
	o.deduction = o.totalPrice * types.Fen(c.Off) / 100
	return nil
}
