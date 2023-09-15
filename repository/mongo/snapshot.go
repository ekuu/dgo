package mongo

import (
	"time"

	repo "github.com/ekuu/dgo/repository"
)

// snapshot 快照
type snapshot[I repo.ID, D any] struct {
	ID     repo.Vid[I]     `bson:"_id"`
	SaveAt time.Time       `bson:"saveAt"`
	Data   Aggregate[I, D] `bson:"data"`
}

func NewSnapshot[I repo.ID, D any]() *snapshot[I, D] {
	return new(snapshot[I, D])
}

func (s *snapshot[I, D]) GetID() repo.Vid[I] {
	return s.ID
}

func (s *snapshot[I, D]) SetID(vid repo.Vid[I]) {
	s.ID = vid
}

func (s *snapshot[I, D]) GetSaveAt() time.Time {
	return s.SaveAt
}

func (s *snapshot[I, D]) SetSaveAt(t time.Time) {
	s.SaveAt = t
}

func (s *snapshot[I, D]) GetAggregate() Aggregate[I, D] {
	return s.Data
}

func (s *snapshot[I, D]) SetAggregate(d Aggregate[I, D]) {
	s.Data = d
}
