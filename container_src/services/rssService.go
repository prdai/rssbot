// Package services provides a service for the RSS service.
package services

import (
	"context"
	"log/slog"
	"sync"

	"github.com/prdai/rssbot/repository"
	"github.com/prdai/rssbot/utils"

	"github.com/mmcdole/gofeed"
	"go.uber.org/dig"
)

type RSSServiceParams struct {
	dig.In

	DBRepository repository.Repository
	RSSParser    *gofeed.Parser
}

type RSSService interface {
	SyncRSSFeeds(rssFeeds []string, ctx context.Context) []string
	getRSSFeed(url string) any
	syncRSSFeed(url string)
}

type rssService struct {
	dbRepository repository.Repository
	rssParser    *gofeed.Parser
}

func (r *rssService) SyncRSSFeeds(rssFeeds []string, ctx context.Context) []string {
	for _, rssFeed := range rssFeeds {
		slog.Info(rssFeed)
		go r.syncRSSFeed(rssFeed)
	}
	return make([]string, 0)
}

func (r *rssService) syncRSSFeed(url string) {
	feedHash := utils.ConvertStringToHash(url)
	feedFetcherChan := make(chan *gofeed.Feed, 1)
	feedRetrivalChan := make(chan *repository.Feed, 1)
	var wg sync.WaitGroup
	wg.Add(2)
	go r.getRSSFeed(url, feedFetcherChan, &wg)
	go r.dbRepository.GetFeed(feedHash, feedRetrivalChan)
	wg.Wait()
	fetchedFeed := <-feedFetcherChan
	retrivedFeed := <-feedRetrivalChan
	if fetchedFeed == nil || retrivedFeed == nil {
		return
	}
}

func (r *rssService) getRSSFeed(url string, feedCollector chan *gofeed.Feed, wg *sync.WaitGroup) {
	defer wg.Done()
	feed, err := r.rssParser.ParseURL(url)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	feedCollector <- feed
}

func NewRSSService(p RSSServiceParams) *rssService {
	return &rssService{dbRepository: p.DBRepository, rssParser: p.RSSParser}
}

func NewRSSParser() *gofeed.Parser {
	return gofeed.NewParser()
}
