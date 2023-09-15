package mongo

import (
	"time"

	dr "github.com/ekuu/dgo/repository"
	repo "github.com/ekuu/dgo/repository"
)

type Aggregate[I repo.ID, D any] interface {
	GetID() I
	SetID(I)
	Content[D]
	GetContent() Content[D]
}

type Content[D any] interface {
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetVersion() uint64
	GetSnapshotSaveAt() time.Time
	GetData() D
	SetCreatedAt(time.Time)
	SetUpdatedAt(time.Time)
	SetVersion(uint64)
	SetSnapshotSaveAt(time.Time)
	SetData(D)
}

type aggregate[I dr.ID, D any] struct {
	ID                 I `bson:"_id"`
	*DefaultContent[D] `bson:",inline"`
}

func NewAggregate[I dr.ID, D any]() Aggregate[I, D] {
	return &aggregate[I, D]{DefaultContent: &DefaultContent[D]{}}
}

func (a *aggregate[I, D]) GetContent() Content[D] {
	return a.DefaultContent
}

func (a *aggregate[I, D]) GetID() I {
	return a.ID
}

func (a *aggregate[I, D]) SetID(id I) {
	a.ID = id
}

type DefaultContent[D any] struct {
	CreatedAt      Time   `bson:"createdAt"`
	UpdatedAt      Time   `bson:"updatedAt"`
	Version        uint64 `bson:"version"`
	Data           D      `bson:",inline"`
	snapshotSaveAt time.Time
}

func (c *DefaultContent[D]) GetCreatedAt() time.Time {
	return c.CreatedAt.Time
}

func (c *DefaultContent[D]) SetCreatedAt(t time.Time) {
	c.CreatedAt = WrapTime(t)
}

func (c *DefaultContent[D]) GetUpdatedAt() time.Time {
	return c.UpdatedAt.Time
}

func (c *DefaultContent[D]) SetUpdatedAt(t time.Time) {
	c.UpdatedAt = WrapTime(t)
}

func (c *DefaultContent[D]) GetVersion() uint64 {
	return c.Version
}

func (c *DefaultContent[D]) SetVersion(version uint64) {
	c.Version = version
}

func (c *DefaultContent[D]) GetSnapshotSaveAt() time.Time {
	return c.snapshotSaveAt
}

func (c *DefaultContent[D]) SetSnapshotSaveAt(snapshotSaveAt time.Time) {
	c.snapshotSaveAt = snapshotSaveAt
}

func (c *DefaultContent[D]) GetData() D {
	return c.Data
}

func (c *DefaultContent[D]) SetData(d D) {
	c.Data = d
}
