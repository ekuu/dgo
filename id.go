package dgo

import (
	"context"

	"github.com/ekuu/dgo/internal"
)

// ID 聚合根ID
type ID string

func NewID(id string) ID {
	return ID(id)
}

func (id ID) IsEmpty() bool {
	return id == ""
}

func (id ID) NotEmpty() bool {
	return !id.IsEmpty()
}

func (id ID) String() string {
	return string(id)
}

func (id ID) isActionTarget() {}

// Vid version and ID
type Vid interface {
	ID() ID
	Version() uint64
}

func NewVid(id ID, version uint64) Vid {
	return NewAggBase(WithAggBaseId(id), WithAggBaseVersion(version))
}

// IDGenerator ID生成接口
type IDGenerator interface {
	GenID(ctx context.Context) (ID, error)
}

type IDGenFunc func(ctx context.Context) (ID, error)

func (f IDGenFunc) GenID(ctx context.Context) (ID, error) {
	return f(ctx)
}

func GenNoHyphenUUID(ctx context.Context) (ID, error) {
	return ID(internal.UUIDNoHyphen()), nil
}
