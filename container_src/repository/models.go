// Package repository provides a repository for the database.
package repository

import "sync"

type Repository interface {
	GetFeed(feedHash string, c chan *Feed, wg *sync.WaitGroup)
	CreateFeed(feedHash string, wg *sync.WaitGroup)
	UpdateFeed(feedHash string, lastItemHash string)
}

type Feed struct {
	FeedHash     string
	LastItemHash string
}
