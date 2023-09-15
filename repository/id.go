package repository

import (
	"strconv"

	"github.com/ekuu/dgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// String 字符串ID
type String string

func (s String) Reverse() dgo.ID {
	return dgo.ID(s)
}

func ParseStringID(id dgo.ID) (String, error) {
	return String(id), nil
}

// U64 uint64 ID
type U64 uint64

func (u U64) Reverse() dgo.ID {
	return dgo.ID(strconv.FormatUint(uint64(u), 10))
}

func ParseU64(id dgo.ID) (U64, error) {
	u, err := strconv.ParseUint(id.String(), 10, 64)
	if err != nil {
		return 0, err
	}
	return U64(u), nil
}

// ObjectID mongodb默认ID
type ObjectID primitive.ObjectID

func (o ObjectID) Reverse() dgo.ID {
	return dgo.ID(primitive.ObjectID(o).Hex())
}

func (id ObjectID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(primitive.ObjectID(id))
}

func (id *ObjectID) UnmarshalBSONValue(bt bsontype.Type, data []byte) error {
	v, rem, ok := bsoncore.ReadObjectID(data)
	if !ok {
		return bsoncore.NewInsufficientBytesError(data, rem)
	}
	*id = ObjectID(v)
	return nil
}

func NewObjectID() ObjectID {
	return ObjectID(primitive.NewObjectID())
}

func ParseObjectID(id dgo.ID) (o ObjectID, err error) {
	v, err := primitive.ObjectIDFromHex(id.String())
	if err != nil {
		return o, err
	}
	return ObjectID(v), nil
}
