package product

import (
	"github.com/ekuu/dgo"
	"github.com/ekuu/dgo/internal/examples/infra/types"
)

//go:generate gogen option -n Product -p _ -r AggBase
type Product struct {
	dgo.AggBase
	name  string
	price types.Fen
}

func (p *Product) Name() string {
	return p.name
}

func (p *Product) Price() types.Fen {
	return p.price
}
