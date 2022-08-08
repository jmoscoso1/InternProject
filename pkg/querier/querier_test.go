package querier

import (
	"testing"
	"context"

	"mymodule/pkg/structs"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/require"
	prom_storage "github.com/prometheus/prometheus/storage"
)

type mockReadClient struct {
	name string
	data int
}


func (r *mockReadClient) Read(ctx context.Context, query *prompb.Query) (*prompb.QueryResult, error) {
	r.data = 5
	r.name = "test"
	return &prompb.QueryResult{}, nil
} 

func Test_Select(t *testing.T) {
	r := &mockReadClient{}

	queryable := NewQueryable(r)
	
	require.NotNil(t, queryable)

	querier, err := queryable.Querier(&structs.EmptyCtx{}, 0, 0)

	require.Nil(t, err)
	require.NotNil(t, querier)


	mockquerier, ok := querier.(Querier) // validates that querier is of type Querier
	require.True(t, ok)

	mockquerier.Select(true, &prom_storage.SelectHints{})

	client, ok := mockquerier.client.(*mockReadClient)
	require.True(t, ok)

	require.Equal(t, client.data, 5)
	require.Equal(t, client.name, "test")
}