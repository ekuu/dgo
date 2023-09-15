package order

import (
	"github.com/ekuu/dgo"
	"github.com/ekuu/dgo/internal/examples/infra/types"
	"github.com/samber/lo"
)

type Item struct {
	productID dgo.ID
	price     types.Fen // 商品单价
	count     uint32
}

func NewItem(productID dgo.ID, price types.Fen, count uint32) *Item {
	return &Item{productID: productID, price: price, count: count}
}

func (e *Item) ProductID() dgo.ID {
	return e.productID
}

func (e *Item) Price() types.Fen {
	return e.price
}

func (e *Item) Count() uint32 {
	return e.count
}

func (e *Item) TotalPrice() types.Fen {
	return types.Fen(e.Count()) * e.Price()
}

type Items []*Item

func (i Items) TotalPrice() types.Fen {
	return lo.SumBy(i, func(v *Item) types.Fen {
		return v.TotalPrice()
	})
}
