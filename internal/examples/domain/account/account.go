package account

import (
	"github.com/ekuu/dgo"
	"github.com/pkg/errors"
)

//go:generate gogen option -n Account -p _ -r AggBase
type Account struct {
	dgo.AggBase
	name    string
	balance uint64
}

func (a *Account) Name() string {
	return a.name
}

func (a *Account) Balance() uint64 {
	return a.balance
}

func (a *Account) Validate() error {
	if a.name == "" {
		return errors.New("账户名不能为空")
	}
	return nil
}
