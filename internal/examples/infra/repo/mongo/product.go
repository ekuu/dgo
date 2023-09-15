package mongo

import (
	"context"

	"github.com/ekuu/dgo/internal/examples/domain/product"

	"github.com/ekuu/dgo/internal/examples/infra/types"

	"github.com/ekuu/dgo"
	dr "github.com/ekuu/dgo/repository"
	dm "github.com/ekuu/dgo/repository/mongo"
	"go.mongodb.org/mongo-driver/mongo"
)

type productRepo struct {
	*dm.Repo[dr.ObjectID, *product.Product, *productPO]
}

type productPO struct {
	Name  string
	Price types.Fen
}

func NewProductRepo(db *mongo.Database) *productRepo {
	convert := func(ctx context.Context, a *product.Product) (*productPO, error) {
		return &productPO{
			Name:  a.Name(),
			Price: a.Price(),
		}, nil
	}
	reverse := func(ctx context.Context, b dgo.AggBase, p *productPO) (*product.Product, error) {
		return product.New(
			b,
			product.WithName(p.Name),
			product.WithPrice(p.Price),
		), nil
	}
	base := dm.NewDefaultRepo[dr.ObjectID, *product.Product, *productPO](
		db,
		"Product",
		convert,
		reverse,
		func() *productPO {
			return new(productPO)
		},
		dr.ParseObjectID,
		dm.WithRepoCloseTransaction[dr.ObjectID, *product.Product, *productPO](true),
	)
	return &productRepo{Repo: base}
}
