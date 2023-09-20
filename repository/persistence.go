package repository

import (
	"time"

	"github.com/ekuu/dgo"
	"google.golang.org/protobuf/encoding/protojson"
)

type ID interface {
	Reverse() dgo.ID
}

type Vid[I ID] interface {
	GetID() I
	SetID(id I)
	GetVersion() uint64
	SetVersion(version uint64)
}

type ParseID[I ID] func(id dgo.ID) (I, error)

type NewVid[I ID] func(id dgo.ID, version uint64) (Vid[I], error)

// Aggregate 仓储层聚合接口,任何仓储实现必须实现此接口
type Aggregate[I ID] interface {
	AggBaseGetter[I]
	AggBaseSetter[I]
}

type AggBaseSetter[I ID] interface {
	SetID(id I)
	SetCreatedAt(t time.Time)
	SetUpdatedAt(t time.Time)
	SetVersion(version uint64)
}

type AggBaseGetter[T ID] interface {
	GetID() T
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetVersion() uint64
}

// ConvertAggBase 从DomainObject转为PersistenceObject
func ConvertAggBase[I ID, P AggBaseSetter[I]](a dgo.AggBase, p P, parseID ParseID[I]) (P, error) {
	id, err := parseID(a.ID())
	if err != nil {
		return p, err
	}
	p.SetID(id)
	p.SetCreatedAt(a.CreatedAt())
	p.SetUpdatedAt(a.UpdatedAt())
	p.SetVersion(a.Version())
	return p, nil
}

// ReverseAggBase 从PersistenceObject转为DomainObject
func ReverseAggBase[I ID](p AggBaseGetter[I]) dgo.AggBase {
	return dgo.NewAggBase(
		dgo.WithAggBaseId(p.GetID().Reverse()),
		dgo.WithAggBaseCreatedAt(p.GetCreatedAt()),
		dgo.WithAggBaseUpdatedAt(p.GetUpdatedAt()),
		dgo.WithAggBaseVersion(p.GetVersion()),
	)
}

type Event[I ID] interface {
	GetID() Vid[I]
	SetID(vid Vid[I])
	GetName() string
	SetName(eventName string)
	GetCreatedAt() time.Time
	SetCreatedAt(t time.Time)
	GetUUID() string
	SetUUID(uuid string)
	GetPayload() []byte
	SetPayload(payload []byte)
}

func ConvertEvent[I ID, E Event[I]](de *dgo.Event, newEvent func() E, newVid NewVid[I]) (E, error) {
	re := newEvent()
	payload, err := protojson.Marshal(de.Payload())
	if err != nil {
		return re, err
	}
	vid, err := newVid(de.AggID(), de.AggVersion())
	if err != nil {
		return re, err
	}
	re.SetName(de.Name())
	re.SetID(vid)
	re.SetCreatedAt(de.CreatedAt())
	re.SetUUID(de.UUID())
	re.SetPayload(payload)
	return re, nil
}
