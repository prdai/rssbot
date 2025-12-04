package services

import (
	"context"
	"sync"

	"github.com/mmcdole/gofeed"
	"github.com/prdai/rssbot/repository"
	"go.uber.org/dig"
)

type RSSServiceParams struct {
	dig.In

	DBRepository repository.Repository
	RSSParser    *gofeed.Parser
}

type RSSService interface {
	SyncRSSFeeds(rssFeeds []string, ctx context.Context) []*NewItems
	getRSSFeed(url string, feedCollector chan *gofeed.Feed, wg *sync.WaitGroup)
	syncRSSFeed(url string, c chan *NewItems)
	captureNewItems(items []*gofeed.Item, wg *sync.WaitGroup, lastItemHash string, newItemsChan chan *NewItems)
}

func NewRSSParser() *gofeed.Parser {
	return gofeed.NewParser()
}

type NewItems struct {
	LatestItemHash string
	Items          []*gofeed.Item
}
