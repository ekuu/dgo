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
		a, err := c.NameExists(ctx, c.Name)
		if err != nil {
			return err
		}
		if a != nil {
			return dgo.NewDuplicate(a)
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

func (c *TransferCmd) BatchEntries() (entries []*dgo.BatchEntry[*Account]) {
	entries = append(
		entries,
		dgo.NewBatchEntryByFunc(
			func(ctx context.Context, a *Account) error {
				a.balance -= c.Amount
				a.AddEvent(&pb.AccountBalanceDecreased{Amount: c.Amount})
				return nil
			},
			dgo.ActionUpdate,
			c.From,
		),
		dgo.NewBatchEntryByFunc(
			func(ctx context.Context, a *Account) error {
				a.balance += c.Amount
				a.AddEvent(&pb.AccountBalanceIncreased{Amount: c.Amount})
				return nil
			},
			dgo.ActionUpdate,
			c.To,
		),
	)
	return
}
