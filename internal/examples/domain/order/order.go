package order

import (
	"fmt"

	"github.com/ekuu/dgo"
	"github.com/ekuu/dgo/internal/examples/infra/types"
)

//go:generate gogen option -n Order -p _ -r AggBase
type Order struct {
	dgo.AggBase
	items      Items
	totalPrice types.Fen
	deduction  types.Fen
}

func (o *Order) AggName() string {
	fmt.Println("这里重写了订单聚合的名字")
	return "OrderRenamed"
}

func (o *Order) Items() Items {
	return o.items
}

func (o *Order) TotalPrice() types.Fen {
	return o.totalPrice
}

func (o *Order) Deduction() types.Fen {
	return o.deduction
}

func (o *Order) PaymentPrice() types.Fen {
	return o.totalPrice - o.deduction
}
