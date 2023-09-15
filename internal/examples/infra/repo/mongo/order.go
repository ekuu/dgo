package mongo

import (
	"context"

	"github.com/samber/lo"

	"github.com/ekuu/dgo/internal/examples/domain/order"

	"github.com/ekuu/dgo/internal/examples/infra/types"

	"github.com/ekuu/dgo"
	dr "github.com/ekuu/dgo/repository"
	dm "github.com/ekuu/dgo/repository/mongo"
	"go.mongodb.org/mongo-driver/mongo"
)

type orderRepo struct {
	*dm.Repo[dr.String, *order.Order, *orderPO]
}

type orderPO struct {
	Items        []orderItem `bson:"items"`
	TotalPrice   types.Fen   `bson:"totalPrice"`
	Deduction    types.Fen   `bson:"deduction"`
	PaymentPrice types.Fen   `bson:"paymentPrice"`
}

type orderItem struct {
	ProductID dgo.ID    `bson:"productID"`
	Price     types.Fen `bson:"price"`
	Count     uint32    `bson:"count"`
}

func (i orderItem) DO() *order.Item {
	return order.NewItem(i.ProductID, i.Price, i.Count)
}

func NewOrderRepo(db *mongo.Database) *orderRepo {
	convert := func(ctx context.Context, o *order.Order) (*orderPO, error) {
		items := lo.Map(o.Items(), func(v *order.Item, index int) orderItem {
			return orderItem{ProductID: v.ProductID(), Price: v.Price(), Count: v.Count()}
		})
		return &orderPO{
			Items:        items,
			TotalPrice:   o.TotalPrice(),
			Deduction:    o.Deduction(),
			PaymentPrice: o.PaymentPrice(),
		}, nil
	}
	reverse := func(ctx context.Context, b dgo.AggBase, p *orderPO) (*order.Order, error) {
		items := lo.Map(p.Items, func(item orderItem, index int) *order.Item {
			return item.DO()
		})
		return order.New(
			b,
			order.WithItems(items),
			order.WithTotalPrice(p.TotalPrice),
			order.WithDeduction(p.Deduction),
		), nil
	}
	base := dm.NewDefaultRepo[dr.String, *order.Order, *orderPO](
		db,
		"Order",
		convert,
		reverse,
		func() *orderPO {
			return new(orderPO)
		},
		dr.ParseStringID,
		dm.WithRepoCloseTransaction[dr.String, *order.Order, *orderPO](true),
	)
	return &orderRepo{Repo: base}
}
