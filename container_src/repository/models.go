// Package repository provides a repository for the database.
package repository

type Repository interface {
	getFeed(feedUrl string)
}

type Feed struct {
	_id string
}
