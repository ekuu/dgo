package mongo

import (
	"time"

	repo "github.com/ekuu/dgo/repository"
)

// event 事件
type event[I repo.ID] struct {
	ID        repo.Vid[I] `bson:"_id"`
	Name      string      `bson:"name"`
	CreatedAt time.Time   `bson:"createdAt"`
	UUID      string      `bson:"uuid"`
	Payload   string      `bson:"payload"`
}

func NewEvent[I repo.ID]() repo.Event[I] {
	return new(event[I])
}

func (e *event[I]) GetID() repo.Vid[I] {
	return e.ID
}

func (e *event[I]) SetID(vid repo.Vid[I]) {
	e.ID = vid
}

func (e *event[I]) GetName() string {
	return e.Name
}

func (e *event[I]) SetName(name string) {
	e.Name = name
}

func (e *event[I]) GetCreatedAt() time.Time {
	return e.CreatedAt
}

func (e *event[I]) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

func (e *event[I]) GetUUID() string {
	return e.UUID
}

func (e *event[I]) SetUUID(uuid string) {
	e.UUID = uuid
}

func (e *event[I]) GetPayload() []byte {
	return []byte(e.Payload)
}

func (e *event[I]) SetPayload(payload []byte) {
	e.Payload = string(payload)
}
