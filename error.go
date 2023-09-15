package dgo

import (
	"errors"
	"fmt"

	pkgerr "github.com/pkg/errors"
)

var (
	ErrIDNil      = errors.New("id is nil")
	ErrNotFound   = errors.New("aggregate not found")
	ErrNotMatched = errors.New("aggregate not matched")
)

func IgnoreNotFound(err error) error {
	if err == nil {
		return nil
	}
	if pkgerr.Is(err, ErrNotFound) {
		return nil
	}
	return err
}

func IgnoreIDNil(err error) error {
	if err == nil || pkgerr.Is(err, ErrIDNil) {
		return nil
	}
	return err
}

type ErrAggCreated[A AggBase] struct {
	a A
}

func NewAggCreated[A AggBase](a A) *ErrAggCreated[A] {
	return &ErrAggCreated[A]{a: a}
}

func (e *ErrAggCreated[A]) Error() string {
	return fmt.Sprintf("aggregate was created, id:%s, createdAt:%s", e.a.ID(), e.a.CreatedAt())
}

func (e *ErrAggCreated[A]) Aggregate() A {
	return e.a
}
