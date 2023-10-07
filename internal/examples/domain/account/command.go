package account

import (
	"context"

	"github.com/ekuu/dgo"
	"github.com/ekuu/dgo/internal/examples/pb"
)

type CreateCmd struct {
	Name       string
	Balance    uint64
	NameExists func(ctx context.Context, name string) (*Account, error)
}

func (c *CreateCmd) Handle(ctx context.Context, a *Account) error {
	a.name = c.Name
	a.balance = c.Balance
	if c.NameExists != nil {
		a2, err := c.NameExists(ctx, c.Name)
		if err != nil {
			return err
		}
		if a2 != nil {
			*a = *a2
			return nil
		}
	}

	//a.AddEvent(
	//	&pb.AccountCreated{
	//		Name:    c.Name,
	//		Balance: c.Balance,
	//	},
	//	dgo.WithEventName("CreatedAssignedInOption"),
	//)
	return nil
}

type UpdateNameCmd struct {
	ID   dgo.ID
	Name string
}

func (c UpdateNameCmd) Handle(ctx context.Context, a *Account) error {
	a.name = c.Name
	a.AddEvent(&pb.AccountNameUpdated{Name: a.name})
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
