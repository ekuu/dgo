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

type ErrDuplicate[A AggBase] struct {
	a A
}

func NewDuplicate[A AggBase](a A) *ErrDuplicate[A] {
	return &ErrDuplicate[A]{a: a}
}

func (e *ErrDuplicate[A]) Error() string {
	return fmt.Sprintf("aggregate duplicate, id:%s, createdAt:%s", e.a.ID(), e.a.CreatedAt())
}

func (e *ErrDuplicate[A]) Aggregate() A {
	return e.a
}
