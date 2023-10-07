package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/ekuu/dgo"
	"github.com/ekuu/dgo/internal/examples/domain/account"
	dr "github.com/ekuu/dgo/repository"
	dm "github.com/ekuu/dgo/repository/mongo"
	"go.mongodb.org/mongo-driver/mongo"
)

type accountRepo struct {
	*dm.Repo[dr.String, *account.Account, *accountPO]
}

type accountPO struct {
	Name    string
	Balance uint64
}

func NewAccountRepo(db *mongo.Database) *accountRepo {
	parseID := dr.ParseStringID
	convert := func(ctx context.Context, a *account.Account) (*accountPO, error) {
		return &accountPO{
			Name:    a.Name(),
			Balance: a.Balance(),
		}, nil
	}
	reverse := func(ctx context.Context, b dgo.AggBase, d *accountPO) (*account.Account, error) {
		return account.New(
			b,
			account.WithName(d.Name),
			account.WithBalance(d.Balance),
		), nil
	}
	base := dm.NewDefaultRepo[dr.String, *account.Account, *accountPO](
		db,
		"Account",
		convert,
		reverse,
		func() *accountPO {
			return new(accountPO)
		},
		parseID,
		dm.WithRepoCloseTransaction[dr.String, *account.Account, *accountPO](true),
	)
	return &accountRepo{Repo: base}
}

func (r accountRepo) NameExists(ctx context.Context, name string) (*account.Account, error) {
	d, err := r.FindDA(ctx, bson.M{"name": name})
	return d, dgo.IgnoreNotFound(err)
}
