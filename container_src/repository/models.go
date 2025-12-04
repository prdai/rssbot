// Package repository provides a repository for the database.
package repository

const RepositoryName = "rssbot"

type Repository interface {
	GetFeed(feedHash string, c chan *Feed)
}

type Feed struct {
	_id          string
	FeedHash     string
	LastItemHash string
}
