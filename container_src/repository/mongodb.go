// Package repository provides a repository for the database.
package repository

import (
	"context"
	"log/slog"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type mongoDBRepository struct {
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
}

func NewMongoDBRepository() (mongoDBRepository, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(os.Getenv("MONGODB_URI")).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(opts)
	if err != nil {
		slog.Error(err.Error())
		return mongoDBRepository{}, err
	}
	db := client.Database(RepositoryName)
	coll := db.Collection(RepositoryName)
	return mongoDBRepository{client: client, db: db, collection: coll}, nil
}

func (m mongoDBRepository) GetFeed(feedHash string, c chan *Feed) {
	filter := bson.D{{"feedHash", feedHash}}
	cursor, err := m.collection.Find(context.TODO(), filter)
	if err != nil {
		slog.Error(err.Error())
		c <- nil
		return
	}
	var results []Feed
	if err = cursor.All(context.TODO(), &results); err != nil {
		slog.Error(err.Error())
		c <- nil
		return
	}
	if len(results) > 1 {
		slog.Error("more than one feed was fetched for the same feedHash")
		c <- nil
		return
	}
	feedResult := results[0]
	c <- &feedResult
}
