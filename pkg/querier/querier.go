package querier

import (
	"context"
	"fmt"

	"github.com/prometheus/prometheus/model/labels"
	prom_storage "github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/remote"
)

type Queryable struct {
	client remote.ReadClient
}

func NewQueryable(c remote.ReadClient) Queryable {
	return Queryable{client: c}
}

func (q Queryable) Querier(ctx context.Context, mint, maxt int64) (prom_storage.Querier, error) {
	return Querier{client: q.client, ctx: ctx, mint: mint, maxt: maxt}, nil
}

type Querier struct {
	ctx        context.Context
	mint, maxt int64
	client     remote.ReadClient
}

func (q Querier) Select(sortSeries bool, hints *prom_storage.SelectHints, matchers ...*labels.Matcher) prom_storage.SeriesSet {
	fmt.Println("Select - RemoteRead")
	query, err := remote.ToQuery(q.mint, q.maxt, matchers, hints)
	if err != nil {
		return prom_storage.ErrSeriesSet(fmt.Errorf("toQuery: %w", err))
	}

	res, err := q.client.Read(q.ctx, query)
	if err != nil {
		return prom_storage.ErrSeriesSet(fmt.Errorf("remote_read: %w", err))
	}
	return remote.FromQueryResult(sortSeries, res)
}

func (q Querier) Close() error {
	return nil
}

func (q Querier) LabelNames(matchers ...*labels.Matcher) ([]string, prom_storage.Warnings, error) {
	return nil, nil, nil
}

func (q Querier) LabelValues(name string, matchers ...*labels.Matcher) ([]string, prom_storage.Warnings, error) {
	return nil, nil, nil
}
