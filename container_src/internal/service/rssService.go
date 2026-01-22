// Package services provides a service for the RSS service.
package services

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/mmcdole/gofeed"
	"github.com/prdai/rssbot/repository"
	"github.com/prdai/rssbot/utils"
)

type rssService struct {
	dbRepository         repository.Repository
	rssParser            *gofeed.Parser
	untrackedFeedMaxItem int
}

func (r *rssService) SyncRSSFeeds(rssFeeds []string, ctx context.Context) []*NewItems {
	rssFeedsNewItemsChan := make(chan *NewItems, len(rssFeeds))
	var wg sync.WaitGroup
	for _, rssFeed := range rssFeeds {
		wg.Add(1)
		go r.syncRSSFeed(rssFeed, rssFeedsNewItemsChan, &wg)
	}
	wg.Wait()
	var rssFeedsNewItems []*NewItems
	for range len(rssFeeds) {
		rssFeedItems := <-rssFeedsNewItemsChan
		rssFeedsNewItems = append(rssFeedsNewItems, rssFeedItems)
	}
	return rssFeedsNewItems
}

func (r *rssService) syncRSSFeed(url string, c chan *NewItems, feedWg *sync.WaitGroup) {
	defer feedWg.Done()
	feedHash := utils.ConvertStringToHash(url)
	feedFetcherChan := make(chan *gofeed.Feed, 1)
	feedRetrivalChan := make(chan *repository.Feed, 1)

	var wg sync.WaitGroup
	wg.Add(1)
	go r.getRSSFeed(url, feedFetcherChan, &wg)
	wg.Wait()
	fetchedFeed := <-feedFetcherChan
	if fetchedFeed == nil {
		return
	}

	wg.Add(1)
	go r.dbRepository.GetFeed(feedHash, feedRetrivalChan, &wg)
	wg.Wait()
	retrivedFeed := <-feedRetrivalChan

	newItemsChan := make(chan *NewItems, len(fetchedFeed.Items))
	if retrivedFeed == nil {
		wg.Add(1)
		retrivedFeed = &repository.Feed{}
		go r.dbRepository.CreateFeed(feedHash, &wg)
	}
	wg.Add(1)
	go r.captureNewItems(fetchedFeed.Items, &wg, retrivedFeed.LastItemHash, newItemsChan)
	wg.Wait()
	newItems := <-newItemsChan
	go r.dbRepository.UpdateFeed(feedHash, newItems.LatestItemHash)
	c <- newItems
}

func (r *rssService) captureNewItems(items []*gofeed.Item, wg *sync.WaitGroup, lastItemHash string, newItemsChan chan *NewItems) {
	defer wg.Done()
	var firstHashString string
	var newItems []*gofeed.Item
	for i, item := range items {
		hashString, err := utils.ConvertObjectToHash(item)
		if err != nil {
			// TODO: create a copy of the gofeed item and then implement String()
			// slog.Error("Unable to Hash Item: %s, Error Raised: %s", string(item), err.Error())
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
	untrackedFeedMaxItem, err := strconv.Atoi(os.Getenv("UNTRACKED_FEED_MAX_ITEMS"))
	if err != nil {
		slog.Error(err.Error())
		panic(err.Error())
	}
	return &rssService{dbRepository: p.DBRepository, rssParser: p.RSSParser, untrackedFeedMaxItem: untrackedFeedMaxItem}
}

func InvokeRssSync(RSSFeeds RSSFeeds, s services.RSSService, ai *clients.AIClient, r *http.Request) error {
	slog.Info("Invoking Sync RSS Feeds")
	newRssFeedsItems := s.SyncRSSFeeds(RSSFeeds.Feeds, r.Context())
	email, err := ai.GenerateEmail(newRssFeedsItems)
	if err != nil {
		slog.Error(err.Error())
		return nil
	}
	clients.SendEmail(email.Title, email.HTMLBody)
	return nil
}
