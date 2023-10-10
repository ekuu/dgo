package app

import (
	"context"

	"github.com/samber/lo"

	"github.com/ekuu/dgo"

	"github.com/ekuu/dgo/internal/examples/infra/dep"

	"github.com/ekuu/dgo/internal/examples/domain/product"
)

func CreateProduct(ctx context.Context, cmd *product.CreateCmd) (*product.Product, error) {
	return dep.ProductSvc().Create(ctx, cmd)
}

func CreateProducts(ctx context.Context, cmds []*product.CreateCmd) ([]*product.Product, error) {
	entries := lo.Map(cmds, func(item *product.CreateCmd, index int) *dgo.BatchEntry[*product.Product] {
		return dgo.NewBatchEntry[*product.Product](item, dgo.ActionCreate, nil)
	})
	return dep.ProductSvc().Batch(ctx, entries)
}
