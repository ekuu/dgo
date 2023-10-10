package app

import (
	"context"

	"github.com/ekuu/dgo/internal/examples/domain/order"
	"github.com/samber/lo"

	"github.com/ekuu/dgo"

	"github.com/ekuu/dgo/internal/examples/domain/product"
	"github.com/ekuu/dgo/internal/examples/infra/dep"
)

func CreateOrder(ctx context.Context, pairs []lo.Entry[dgo.ID, uint32], off uint8) (*order.Order, error) {
	// 提取产品ID
	productIDs := lo.Map(
		pairs,
		func(item lo.Entry[dgo.ID, uint32], index int) dgo.ID {
			return item.Key
		},
	)

	// 获取产品明细
	products, err := dep.ProductSvc().List(ctx, productIDs...)
	if err != nil {
		return nil, err
	}

	// 创建订单条目
	m := lo.FromEntries(pairs)
	items := lo.Map(
		products,
		func(p *product.Product, index int) *order.Item {
			return order.NewItem(p.ID(), p.Price(), m[p.ID()])
		},
	)

	// 创建订单
	return dep.OrderSvc().Create(ctx, &order.CreateCmd{Items: items, Off: off})
}
