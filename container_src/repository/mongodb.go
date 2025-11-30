// Package repository provides a repository for the database.
package repository

import "os"

type mongoDBRepository struct {
	mongodbURI string
}

func NewMongoDBRepository() mongoDBRepository {
	return mongoDBRepository{mongodbURI: os.Getenv("MONGODB_URI")}
}

func (m *mongoDBRepository) getName() string {
	return "mongodb"
}
