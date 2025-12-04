// Package repository provides a repository for the database.
package repository

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

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

func (m mongoDBRepository) GetFeed(feedHash string, c chan *Feed, wg *sync.WaitGroup) {
	defer wg.Done()
	filter := bson.D{bson.E{Key: "feedHash", Value: feedHash}}
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
	if len(results) == 0 {
		slog.Error("no feed found with the provided hash")
		c <- nil
		return
	}
	feedResult := results[0]
	c <- &feedResult
}

func (m mongoDBRepository) CreateFeed(feedHash string, wg *sync.WaitGroup) {
	defer wg.Done()
	doc := Feed{FeedHash: feedHash}
	result, err := m.collection.InsertOne(context.TODO(), doc)
	if err != nil {
		slog.Error(err.Error())
	}
	fmt.Printf("%+v\n", *result)
}

func (m mongoDBRepository) UpdateFeed(feedHash string, lastItemHash string) {
	filter := bson.D{bson.E{Key: "feedHash", Value: feedHash}}
	replacement := bson.D{bson.E{Key: "feedHash", Value: feedHash}, bson.E{Key: "lastItemHash", Value: lastItemHash}}
	_, err := m.collection.ReplaceOne(context.TODO(), filter, replacement)
	if err != nil {
		slog.Error(err.Error())
	}
}
