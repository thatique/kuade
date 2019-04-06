package db

import (
	"github.com/thatique/kuade/api/v1"
	bsonp "go.mongodb.org/mongo-driver/bson/primitive"
)

func ToMongoObjectID(object v1.ObjectID) bsonp.ObjectID {
	var b [12]byte
	copy(b[:], object[:])
	return b
}

func FromMongoObjectID(object bsonp.ObjectID) v1.ObjectID {
	var b [12]byte
	copy(b[:], object[:])
	return b
}
