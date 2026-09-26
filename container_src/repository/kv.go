// Package repository provides a repository for the feed state.
package repository

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/kv"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/shared"
)

type kvRepository struct {
	client      *cloudflare.Client
	accountID   string
	namespaceID string
}

func NewKVRepository() (kvRepository, error) {
	client := cloudflare.NewClient(option.WithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN")))
	return kvRepository{
		client:      client,
		accountID:   os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
		namespaceID: os.Getenv("KV_NAMESPACE_ID"),
	}, nil
}

func (r kvRepository) GetFeed(feedHash string, c chan *Feed, wg *sync.WaitGroup) {
	defer wg.Done()
	res, err := r.client.KV.Namespaces.Values.Get(context.TODO(), r.namespaceID, feedHash, kv.NamespaceValueGetParams{
		AccountID: cloudflare.F(r.accountID),
	})
	if err != nil {
		var apiErr *kv.Error
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			c <- nil
			return
		}
		slog.Error(err.Error())
		c <- nil
		return
	}
	defer res.Body.Close()
	value, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error(err.Error())
		c <- nil
		return
	}
	c <- &Feed{FeedHash: feedHash, LastItemHash: string(value)}
}

func (r kvRepository) CreateFeed(feedHash string, wg *sync.WaitGroup) {
	defer wg.Done()
	r.put(feedHash, "")
}

func (r kvRepository) UpdateFeed(feedHash string, lastItemHash string) {
	r.put(feedHash, lastItemHash)
}

func (r kvRepository) put(feedHash string, lastItemHash string) {
	_, err := r.client.KV.Namespaces.Values.Update(context.TODO(), r.namespaceID, feedHash, kv.NamespaceValueUpdateParams{
		AccountID: cloudflare.F(r.accountID),
		Value:     cloudflare.F[kv.NamespaceValueUpdateParamsValueUnion](shared.UnionString(lastItemHash)),
	})
	if err != nil {
		slog.Error(err.Error())
	}
}
