package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Student struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
	FirstName    string             `json:"firstName" bson:"firstName"`
	LastName     string             `json:"lastName" bson:"lastName"`
	PhoneNumber  string             `json:"phoneNumber" bson:"phoneNumber"`
	Courses      []Course           `json:"courses" bson:"courses"`
	ChatId       int64              `json:"chatId" bson:"chatId"`
	State        string             `json:"state" bson:"state"`
	IsRegistered bool               `json:"isRegistered" bson:"isRegistered"`
}
