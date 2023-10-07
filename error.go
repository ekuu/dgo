package dgo

import (
	"errors"

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
