// Package repository provides a repository for the database.
package repository

import (
	"log/slog"
	"os"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type mongoDBRepository struct {
	client *mongo.Client
}

func NewMongoDBRepository() (mongoDBRepository, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(os.Getenv("MONGODB_URI")).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(opts)
	if err != nil {
		slog.Error(err.Error())
		return mongoDBRepository{}, err
	}
	return mongoDBRepository{client: client}, nil
}

func (m mongoDBRepository) getFeed(feedUrl string) {}
