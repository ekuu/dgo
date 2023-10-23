package dgo

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
)

// AggBase 聚合基础接口
type AggBase interface {
	ID() ID
	Now() time.Time
	CreatedAt() time.Time
	UpdatedAt() time.Time
	Version() uint64
	OriginalVersion() uint64
	IsNew() bool
	AddEvent(payload proto.Message, opts ...EventOption)

	base() *aggBase
	setID(id ID)
	isDirty() bool
	isActionTarget()
	getEvents() Events
}

// aggBase 定义聚合通用的基础信息
//
//go:generate gogen option -n aggBase -i versionDelta,events --with-init
type aggBase struct {
	id           ID
	createdAt    time.Time
	updatedAt    time.Time
	version      uint64
	versionDelta uint64
	now          time.Time
	events       Events
	changed      bool
	dryRun       bool
}

func (b *aggBase) init() {
	now := time.Now()
	if b.CreatedAt().IsZero() || b.CreatedAt().Unix() == 0 {
		b.createdAt = now
	}
	if b.UpdatedAt().IsZero() || b.UpdatedAt().Unix() == 0 {
		b.updatedAt = now
	}
	if b.now.IsZero() || b.now.Unix() == 0 {
		b.now = now
	}
}

func (b *aggBase) base() *aggBase {
	return b
}

func (b *aggBase) ID() ID {
	return b.id
}

func (b *aggBase) setID(id ID) {
	b.id = id
	for i, _ := range b.events {
		b.events[i].aggID = id
	}
}

func (b *aggBase) Now() time.Time {
	return b.now
}

func (b *aggBase) Version() uint64 {
	return b.version + b.versionDelta
}

func (b *aggBase) OriginalVersion() uint64 {
	return b.version
}

func (b *aggBase) CreatedAt() time.Time {
	return b.createdAt
}

func (b *aggBase) UpdatedAt() time.Time {
	return b.updatedAt
}

func (b *aggBase) IsNew() bool {
	return b.OriginalVersion() == 0
}

func (b *aggBase) AddEvent(payload proto.Message, opts ...EventOption) {
	e := NewEvent(payload, opts...)
	e.aggID = b.ID()
	b.events = append(b.events, e)
}

func (b *aggBase) tempCleanEvents(fn func(events Events)) {
	events := b.events
	b.events = nil
	fn(events)
	b.events = events
}

func (b *aggBase) getEvents() Events {
	return b.events
}

func (b *aggBase) isDirty() bool {
	return b.changed || b.IsNew()
}

func (b *aggBase) incrVersion() uint64 {
	b.versionDelta++
	return b.Version()
}

func (b *aggBase) completeEvents(v AggBase) {
	if len(b.events) == 0 {
		if b.isDirty() {
			b.incrVersion()
		}
		return
	}
	aggName := getAggName(v)
	topic := GenDefaultTopic(aggName)
	if t := getTopicName(v); t != "" {
		topic = t
	}
	for i, e := range b.events {
		b.events[i].aggName = aggName
		b.events[i].aggVersion = b.incrVersion()
		if e.Topic() == "" {
			b.events[i].topic = topic
		}
	}
}

func (b *aggBase) isActionTarget() {}

var GenDefaultTopic = func(aggName string) string {
	return fmt.Sprintf("%sEvent", aggName)
}
