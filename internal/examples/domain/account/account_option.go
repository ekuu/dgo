// Code generated by "gogen option -n Account -p _ -r AggBase"; DO NOT EDIT.

package account

import (
	"fmt"
	"github.com/ekuu/dgo"
)

// New constructor
func New(aggBase dgo.AggBase, _opts ...Option) *Account {
	_a := new(Account)

	_a.AggBase = aggBase

	_a.SetOptions(_opts...)

	return _a
}

// Option option interface
type Option interface {
	apply(*Account)
}

// Option option function
type optionFunc func(*Account)

func (f optionFunc) apply(_a *Account) {
	f(_a)
}

func (_a *Account) SetOptions(_opts ...Option) *Account {
	for _, _opt := range _opts {
		_opt.apply(_a)
	}
	return _a
}

func SkipOption() Option {
	return optionFunc(func(_a *Account) {
		return
	})
}

func WithOptions(o *options) Option {
	return optionFunc(func(_a *Account) {
		_a.SetOptions(o.opts...)
	})
}

// options options struct
type options struct {
	opts []Option
}

// NewOptions new options struct
func NewOptions() *options {
	return new(options)
}

func (_o *options) Options() []Option {
	return _o.opts
}

func (_o *options) Append(_opts ...Option) *options {
	_o.opts = append(_o.opts, _opts...)
	return _o
}

// Name name option of Account
func (_o *options) Name(name string) *options {
	_o.opts = append(_o.opts, WithName(name))
	return _o
}

// Balance balance option of Account
func (_o *options) Balance(balance uint64) *options {
	_o.opts = append(_o.opts, WithBalance(balance))
	return _o
}

// WithName name option of Account
func WithName(name string) Option {
	return optionFunc(func(_a *Account) {
		_a.name = name
	})
}

// WithBalance balance option of Account
func WithBalance(balance uint64) Option {
	return optionFunc(func(_a *Account) {
		_a.balance = balance
	})
}

func PrintOptions(packageName string) {
	opts := []string{
		"WithName()",
		"WithBalance()",
	}
	if packageName == "" {
		fmt.Printf("opts := []Option{ \n")
		for _, v := range opts {
			fmt.Printf("	%s,\n", v)
		}
	} else {
		fmt.Printf("opts := []%s.Option{ \n", packageName)
		for _, v := range opts {
			fmt.Printf("	%s.%s,\n", packageName, v)
		}
	}
	fmt.Println("}")
}