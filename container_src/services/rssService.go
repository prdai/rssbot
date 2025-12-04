// Package services provides a service for the RSS service.
package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/prdai/rssbot/repository"
	"github.com/prdai/rssbot/utils"

	"github.com/mmcdole/gofeed"
)

type rssService struct {
	dbRepository         repository.Repository
	rssParser            *gofeed.Parser
	untrackedFeedMaxItem int `env:"UNTRACKED_FEED_MAX_ITEMS"`
}

func (r *rssService) SyncRSSFeeds(rssFeeds []string, ctx context.Context) []string {
	rssFeedsNewItemsChan := make(chan *NewItems, len(rssFeeds))
	for _, rssFeed := range rssFeeds {
		slog.Info(rssFeed)
		go r.syncRSSFeed(rssFeed, rssFeedsNewItemsChan)
	}
	rssFeedItems := <-rssFeedsNewItemsChan
	fmt.Printf("%+v\n", rssFeedItems)
	return make([]string, 0)
}

func (r *rssService) syncRSSFeed(url string, c chan *NewItems) {
	feedHash := utils.ConvertStringToHash(url)
	feedFetcherChan := make(chan *gofeed.Feed, 1)
	feedRetrivalChan := make(chan *repository.Feed, 1)
	var wg sync.WaitGroup
	wg.Add(2)
	go r.getRSSFeed(url, feedFetcherChan, &wg)
	go r.dbRepository.GetFeed(feedHash, feedRetrivalChan, &wg)
	wg.Wait()
	fetchedFeed := <-feedFetcherChan
	retrivedFeed := <-feedRetrivalChan
	if fetchedFeed == nil {
		return
	}
	newItemsChan := make(chan *NewItems, len(fetchedFeed.Items))
	wg.Add(2)
	if retrivedFeed == nil {
		retrivedFeed = &repository.Feed{}
		go r.dbRepository.CreateFeed(feedHash, &wg)
	}
	go r.captureNewItems(fetchedFeed.Items, &wg, retrivedFeed.LastItemHash, newItemsChan)
	wg.Wait()
	newItems := <-newItemsChan
	c <- newItems
}

func (r *rssService) captureNewItems(items []*gofeed.Item, wg *sync.WaitGroup, lastItemHash string, newItemsChan chan *NewItems) {
	defer wg.Done()
	var firstHashString string
	var newItems []*gofeed.Item
	for i, item := range items {
		hashString, error := utils.ConvertObjectToHash(item)
		if error != nil {
			continue
		}
		if i == 0 {
			firstHashString = hashString
		}
		if lastItemHash != "" && lastItemHash == hashString {
			break
		}
		newItems = append(newItems, item)
		if lastItemHash == "" && len(newItems) >= r.untrackedFeedMaxItem {
			break
		}
	}
	newItemsChan <- &NewItems{Items: newItems, LatestItemHash: firstHashString}
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
