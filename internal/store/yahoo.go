package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/IsakNorberg/argus/internal/database"
	"github.com/IsakNorberg/argus/internal/sources/yahoo"
	"github.com/IsakNorberg/argus/pkg/models"
)

type YahooStore struct {
	db     database.DB
	client *yahoo.Client
}

func NewYahooStore(db database.DB, client *yahoo.Client) *YahooStore {
	return &YahooStore{
		db:     db,
		client: client,
	}
}

type PriceStats struct {
	Fetched int
	Stored  int
	Errors  int
	Elapsed time.Duration
}

func (s *YahooStore) FetchAndSaveQuotes(ctx context.Context, tickers []string) (*PriceStats, error) {
	start := time.Now()
	stats := &PriceStats{}

	if len(tickers) == 0 {
		return nil, fmt.Errorf("inga tickers angivna")
	}

	// Batch: ticker→priceQuotes
	quotes, err := s.client.FetchQuote(ctx, tickers)
	if err != nil {
		return nil, fmt.Errorf("hämta quotes från Yahoo: %w", err)
	}

	now := time.Now()
	today := now.Format("2006-01-02")
	for _, q := range quotes {
		argusID := fmt.Sprintf("Q-%s-%s-%s", strings.ToUpper(q.Ticker), q.Source, today)
		mq := &models.PriceQuote{
			ArgusID:       argusID,
			Ticker:        strings.ToUpper(q.Ticker),
			Price:         q.Price,
			Currency:      q.Currency,
			MarketCap:     q.MarketCap,
			Volume:        q.Volume,
			ChangePercent: q.ChangePercent,
			Source:        q.Source,
			QuoteDate:     now,
		}

		if err := s.db.UpsertPriceQuote(ctx, mq); err != nil {
			stats.Errors++
			continue
		}
		stats.Stored++
	}

	stats.Fetched = len(tickers) - stats.Errors
	stats.Elapsed = time.Since(start)
	return stats, nil
}
