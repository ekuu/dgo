package dgo

import (
	"reflect"

	"github.com/ekuu/dgo/internal"
)

// DryRunner dry runner
type DryRunner interface {
	DryRun() bool
}

// Validator validator
type Validator interface {
	Validate() error
}

// AggConstructor 构造器接口
type AggConstructor[A AggBase] interface {
	NewAggregate() A
}

// MultiDocuments 该接口应由聚合实现
// 如果聚合的数据存储分布在mysql的多个表中，或者mongodb的多个集合中，或者redis的多个key中的类似情况
// 那么该聚合应该实现此接口，以决定在进行仓储操作时自动区分是否启用事务
type MultiDocuments interface {
	IsMultiDocuments()
}

// AggNamer 实现此接口，则该聚合的名称为AggName()的返回值
type AggNamer interface {
	AggName() string
}

func getAggName(v any) string {
	if v == nil {
		panic("getAggName: value is nil")
	}
	if n, ok := v.(AggNamer); ok {
		return n.AggName()
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return internal.FirstUpper(t.Name())
}

// EventNamer 实现此接口，则事件名称为EventName()的返回值
type EventNamer interface {
	EventName() string
}

func getEventName(v any) string {
	if v == nil {
		panic("getEventName: value is nil")
	}
	if n, ok := v.(EventNamer); ok {
		return n.EventName()
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return internal.FirstUpper(t.Name())
}

// TopicNamer 实现此接口，则消息的topic值为TopicName()的返回值
type TopicNamer interface {
	TopicName() string
}

func getTopicName(v any) string {
	if v == nil {
		panic("getTopicName: value is nil")
	}
	if n, ok := v.(TopicNamer); ok {
		return n.TopicName()
	}
	return ""
}
