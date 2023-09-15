package dgo

import (
	"strings"
	"time"

	"github.com/ekuu/dgo/internal"
	"github.com/ekuu/dgo/pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Events 事件切片
type Events []*Event

// Event 事件
//
//go:generate gogen option -n Event -r payload -s name,topic,createdAt,uuid --with-init
type Event struct {
	name       string
	topic      string
	createdAt  time.Time
	uuid       string
	aggID      ID
	aggName    string
	aggVersion uint64
	payload    proto.Message
}

func (e *Event) init() {
	if e.name == "" {
		e.name = getEventName(e.payload)
	}
	if e.createdAt.IsZero() || e.createdAt.Unix() == 0 {
		e.createdAt = time.Now().Local()
	}
	if e.uuid == "" {
		e.uuid = internal.UUIDNoHyphen()
	}
}

func (e *Event) Name() string {
	return strings.Replace(e.name, e.AggName(), "", 1)
}

func (e *Event) Topic() string {
	return e.topic
}

func (e *Event) AggID() ID {
	return e.aggID
}

func (e *Event) AggName() string {
	return e.aggName
}

func (e *Event) AggVersion() uint64 {
	return e.aggVersion
}

func (e *Event) CreatedAt() time.Time {
	return e.createdAt
}

func (e *Event) UUID() string {
	return e.uuid
}

func (e *Event) Payload() proto.Message {
	return e.payload
}

func (e *Event) ProtoMessage() (*pb.Event, error) {
	payload, err := anypb.New(e.Payload())
	if err != nil {
		return nil, err
	}
	return &pb.Event{
		Name:       e.Name(),
		AggId:      e.AggID().String(),
		AggName:    e.AggName(),
		AggVersion: e.AggVersion(),
		CreatedAt:  timestamppb.New(e.CreatedAt()),
		Uuid:       e.UUID(),
		Payload:    payload,
	}, nil
}

func (e *Event) PublishRequest() (*pb.PublishRequest, error) {
	if v, err := e.ProtoMessage(); err != nil {
		return nil, err
	} else {
		return &pb.PublishRequest{Topic: e.Topic(), Event: v}, nil
	}
}
