package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Course struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	LessonTime  string             `json:"lessonTime" bson:"lessonTime"`
	Price       int                `json:"price" bson:"price"`
	Discount    int                `json:"discount" bson:"discount"`
	StartsDate  time.Time          `json:"startsDate" bson:"startsDate"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
}
