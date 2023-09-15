package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type Time struct {
	time.Time
}

func WrapTime(t time.Time) Time {
	return Time{Time: t.Truncate(time.Millisecond)}
}

func (t *Time) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(t.Time)
}

func (t *Time) UnmarshalBSONValue(bt bsontype.Type, data []byte) error {
	dt, rem, ok := bsoncore.ReadDateTime(data)
	if !ok {
		return bsoncore.NewInsufficientBytesError(data, rem)
	}
	t.Time = bsonx.DateTime(dt).Time().Local()
	return nil
}
