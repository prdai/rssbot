// Package repository provides a repository for the database.
package repository

import (
	"log/slog"
	"os"
)

type mongoDBRepository struct {
	mongodbURI string
}

func NewMongoDBRepository() mongoDBRepository {
	slog.Info("Creating New MongoDB Repository")
	return mongoDBRepository{mongodbURI: os.Getenv("MONGODB_URI")}
}

func (m mongoDBRepository) getName() string {
	return "mongodb"
}
