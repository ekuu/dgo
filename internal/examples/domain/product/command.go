package product

import (
	"context"

	"github.com/ekuu/dgo/internal/examples/infra/types"

	"github.com/ekuu/dgo/internal/examples/pb"

	"github.com/pkg/errors"
)

type CreateCmd struct {
	Name  string
	Price types.Fen
}

func (c CreateCmd) Handle(ctx context.Context, a *Product) error {
	if c.Name == "" {
		return errors.New("product name is empty")
	}

	a.name = c.Name
	a.price = c.Price

	a.AddEvent(&pb.ProductCreated{
		Name:   c.Name,
		Uint64: uint64(c.Price),
	})

	return nil
}
