package pusher

import (
	"context"
	"testing"
	"mymodule/pkg/structs"

	"github.com/cortexproject/cortex/pkg/cortexpb"
	"github.com/stretchr/testify/require"
)

type mockWriteClient struct {
	name string
	data int
}

func (w *mockWriteClient) Store(context.Context, []byte) error {
	w.name = "test"
	w.data = 5
	return nil
}

func (w *mockWriteClient) Name() string {
	return ""
}
func (w *mockWriteClient) Endpoint() string {
	return ""
}

func Test_Push(t *testing.T) {
	w := &mockWriteClient{}

	require.Equal(t, w.data, 0)
	require.Equal(t, w.name, "")

	Pusher := NewPusher(w)
	require.NotNil(t, Pusher)

	_, err := Pusher.Push(&structs.EmptyCtx{}, &cortexpb.WriteRequest{})
	require.NoError(t, err)

	client, ok := Pusher.client.(*mockWriteClient); // validates that client is of type *mockWriteClient

	require.True(t, ok)
	require.Equal(t, client.data, 5)
	require.Equal(t, client.name, "test")
}