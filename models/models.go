package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id     primitive.ObjectID `bson:"_id"`
	Name   string             `bson: "name"`
	Gender string             `bson: "gender"`
	Age    int                `bson: "age"`
}
