package controller

import (
	"context"
	"fajr-acceptance/internal/database"
	"fajr-acceptance/internal/handler/apierr"
	"fajr-acceptance/internal/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type CourseController struct {
	db *database.MongoDBClient
}

const (
	CourseCollection = "courses"
)

type CourseReq struct {
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
	LessonTime  string    `json:"lessonTime" bson:"lessonTime"`
	Price       int       `json:"price" bson:"price"`
	Discount    int       `json:"discount" bson:"discount"`
	StartsDate  time.Time `json:"startsDate" bson:"startsDate"`
}

func NewCourseController(db *database.MongoDBClient) *CourseController {
	return &CourseController{db: db}
}

func (cc *CourseController) GetCourses(gctx *gin.Context) (any, error) {
	coll := cc.db.GetCollection(CourseCollection)

	cursor, err := coll.Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"createdAt": -1}))

	if err != nil {
		return nil, err
	}
	var courses []models.Course
	if err = cursor.All(context.Background(), &courses); err != nil {
		return nil, err
	}
	return courses, nil
}

func (cc *CourseController) AddCourse(gctx *gin.Context) (any, error) {
	coll := cc.db.GetCollection(CourseCollection)
	var req CourseReq
	if err := gctx.ShouldBind(&req); err != nil {
		return nil, apierr.ErrInvalidRequest.WithMessage(err.Error())
	}

	course := models.Course{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Discount:    req.Discount,
		LessonTime:  req.LessonTime,
		StartsDate:  req.StartsDate,
		CreatedAt:   time.Now(),
	}
	res, err := coll.InsertOne(context.TODO(), course)
	if err != nil {
		return nil, err
	}

	var insertedCourse models.Course

	if err = coll.FindOne(context.Background(), bson.M{"_id": res.InsertedID}).Decode(&insertedCourse); err != nil {
		return nil, err
	}

	return insertedCourse, nil
}

func (cc *CourseController) UpdateCourse(gctx *gin.Context) (any, error) {
	coll := cc.db.GetCollection(CourseCollection)
	courseById, err := findCourseById(gctx, coll)
	filter := bson.D{{"_id", courseById.ID}}
	var req CourseReq
	if err := gctx.ShouldBind(&req); err != nil {
		return nil, err
	}
	courseById.Name = req.Name
	courseById.Description = req.Description
	courseById.Price = req.Price
	courseById.StartsDate = req.StartsDate
	courseById.LessonTime = req.LessonTime
	courseById.Discount = req.Discount

	update := bson.M{
		"$set": courseById,
	}

	var result models.Course

	err = coll.FindOneAndUpdate(context.Background(), filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&result)
	if err != nil {
		return nil, err
	}
	return gin.H{"message": "Course updated!", "data": result}, nil
}

func (cc *CourseController) DeleteCourse(gctx *gin.Context) (interface{}, error) {
	coll := cc.db.GetCollection(CourseCollection)
	courseById, err := findCourseById(gctx, coll)
	_, err = cc.db.GetCollection(CourseCollection).DeleteOne(context.Background(), bson.M{"_id": courseById.ID})
	if err != nil {
		return nil, err
	}
	return courseById, nil

}

func findCourseById(gctx *gin.Context, coll *mongo.Collection) (*models.Course, error) {
	id, _ := primitive.ObjectIDFromHex(gctx.Param("id"))

	filter := bson.D{{"_id", id}}

	var courseById models.Course
	err := coll.FindOne(context.Background(), filter).Decode(&courseById)
	if err != nil {

		return nil, apierr.New(409, "409", "Bunday idlik client mavjud emas!")
	}
	return &courseById, nil
}
