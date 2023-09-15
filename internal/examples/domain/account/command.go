package account

import (
	"context"

	"github.com/ekuu/dgo/internal/examples/pb"

	"github.com/ekuu/dgo"
)

type CreateCmd struct {
	Name    string
	Balance uint64
}

func (c *CreateCmd) Handle(ctx context.Context, a *Account) error {
	a.name = c.Name
	a.balance = c.Balance
	a.AddEvent(
		&pb.AccountCreated{
			Name:    c.Name,
			Balance: c.Balance,
		},
		dgo.WithEventName("CreatedAssignedInOption"),
	)
	return nil
}

type UpdateNameCmd struct {
	Name string
}

func (c UpdateNameCmd) Handle(ctx context.Context, a *Account) error {
	a.name = c.Name
	return nil
}

type TransferCmd struct {
	From   dgo.ID
	To     dgo.ID
	Amount uint64
}

func NewTransferCmd(from dgo.ID, to dgo.ID, amount uint64) *TransferCmd {
	return &TransferCmd{From: from, To: to, Amount: amount}
}

func (t TransferCmd) BatchHandle(ctx context.Context) (entries []dgo.BatchEntry[*Account], err error) {
	entries = append(
		entries,
		dgo.NewBatchEntryByFunc(
			func(ctx context.Context, a *Account) error {
				a.balance -= t.Amount
				a.AddEvent(&pb.AccountBalanceDecreased{Amount: t.Amount})
				return nil
			},
			t.From,
		),
		dgo.NewBatchEntryByFunc(
			func(ctx context.Context, a *Account) error {
				a.balance += t.Amount
				a.AddEvent(&pb.AccountBalanceIncreased{Amount: t.Amount})
				return nil
			},
			t.To,
		),
	)
	return
}
