// Package services provides a service for the RSS service.
package services

import (
	"fmt"
	"log/slog"

	"github.com/prdai/rssbot/repository"

	"github.com/mmcdole/gofeed"
	"go.uber.org/dig"
)

type RSSServiceParams struct {
	dig.In

	dbRepository repository.Repository
	rssParser    gofeed.Parser
}

type RSSService interface {
	SyncRSSFeeds(rssFeeds []string) []string
	getRSSFeed(url string) any
}

type rssService struct {
	dbRepository repository.Repository
	rssParser    gofeed.Parser
}

func (r *rssService) SyncRSSFeeds(rssFeeds []string) []string {
	for _, rssFeed := range rssFeeds {
		slog.Info(rssFeed)
		go r.getRSSFeed(rssFeed)
	}
	return make([]string, 0)
}

func (r *rssService) getRSSFeed(url string) any {
	feed, err := r.rssParser.ParseURL(url)
	if err != nil {
		return nil
	}
	fmt.Println(feed)
	return "test"
}

func NewRSSService(params RSSServiceParams) RSSService {
	return &rssService{dbRepository: params.dbRepository, rssParser: params.rssParser}
}

func NewRSSParser() gofeed.Parser {
	return *gofeed.NewParser()
}
