package database

import (
	"context"
	"fajr-acceptance/internal/config"
	"fajr-acceptance/internal/models"
	"fajr-acceptance/pkg/utils/authutil"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type MongoDBClient struct {
	Client *mongo.Client
}

func NewMongoDb(cfg *config.Config, logger *logrus.Logger) (*MongoDBClient, error) {
	logger.Info(cfg.Db.DataSourceName)
	clientOptions := options.Client().ApplyURI(cfg.Db.DataSourceName)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		return nil, err
	}

	loadInitialData(client.Database("fajr_academy"))
	return &MongoDBClient{
		Client: client,
	}, nil
}

func (m *MongoDBClient) Close() error {
	err := m.Client.Disconnect(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (m *MongoDBClient) GetCollection(collectionName string) *mongo.Collection {
	return m.Client.Database("fajr_academy").Collection(collectionName)
}

func loadInitialData(db *mongo.Database) {
	password, _ := authutil.EncodePassword("starlight@99", 0)
	admin := models.User{
		FirstName: "JAKHONGIR",
		LastName:  "Egamberdiyev",
		Username:  "jakhon99dev",
		Email:     "jakhon99dev@gmail.com",
		Password:  password,
	}
	admin.CreatedAt = time.Now()
	admin.UpdatedAt = admin.CreatedAt
	filter := bson.D{
		{"username", admin.Username},
	}
	var user struct{}
	collection := db.Collection("users")
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != mongo.ErrNoDocuments {
		return
	} else if err == nil {
		log.Println(err)
		return
	}
	_, err = collection.InsertOne(context.Background(), admin)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Admin yaratildi.")
}
