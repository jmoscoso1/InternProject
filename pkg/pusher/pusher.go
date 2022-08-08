package pusher

import (
	"context"
	"fmt"
	
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"

	"github.com/cortexproject/cortex/pkg/cortexpb"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
)

type Pusher struct {
	client remote.WriteClient
}

func (p Pusher) Push(ctx context.Context, req *cortexpb.WriteRequest) (*cortexpb.WriteResponse, error) {
	fmt.Println("Push - RemoteWrite")
	var (
		pBuf = proto.NewBuffer(nil)
		buf  []byte
	)
	writeReq, _, err := buildWriteRequest(req, pBuf, buf)
	if err != nil {
		return nil, err
	}

	err = p.client.Store(ctx, writeReq)

	return &cortexpb.WriteResponse{}, err
}

func NewPusher(c remote.WriteClient) Pusher {
	return Pusher{client: c}
}

func buildWriteRequest(cortexReq *cortexpb.WriteRequest, pBuf *proto.Buffer, buf []byte) ([]byte, int64, error) {
	var highest int64
	for _, ts := range cortexReq.Timeseries {
		// At the moment we only ever append a TimeSeries with a single sample or exemplar in it.
		if len(ts.Samples) > 0 && ts.Samples[0].TimestampMs > highest {
			highest = ts.Samples[0].TimestampMs
		}
		if len(ts.Exemplars) > 0 && ts.Exemplars[0].TimestampMs > highest {
			highest = ts.Exemplars[0].TimestampMs
		}
	}

	req := toPromWriteRequest(cortexReq)
	if pBuf == nil {
		pBuf = proto.NewBuffer(nil) // For convenience in tests. Not efficient.
	} else {
		pBuf.Reset()
	}
	err := pBuf.Marshal(req)
	if err != nil {
		return nil, highest, err
	}

	// snappy uses len() to see if it needs to allocate a new slice. Make the
	// buffer as long as possible.
	if buf != nil {
		buf = buf[0:cap(buf)]
	}
	compressed := snappy.Encode(buf, pBuf.Bytes())
	return compressed, highest, nil
}

func toPromWriteRequest(r *cortexpb.WriteRequest) *prompb.WriteRequest {
	req := &prompb.WriteRequest{
		Timeseries: []prompb.TimeSeries{},
		Metadata:   []prompb.MetricMetadata{},
	}

	for _, m := range r.Metadata {
		req.Metadata = append(req.Metadata, prompb.MetricMetadata{
			Type:             prompb.MetricMetadata_MetricType(m.Type),
			MetricFamilyName: m.MetricFamilyName,
			Help:             m.Help,
			Unit:             m.Unit,
		})
	}

	for _, t := range r.Timeseries {
		labels := []prompb.Label{}
		samples := []prompb.Sample{}

		for _, l := range t.Labels {
			labels = append(labels, prompb.Label{
				Name:  l.Name,
				Value: l.Value,
			})
		}

		for _, s := range t.Samples {
			samples = append(samples, prompb.Sample{
				Value:     s.Value,
				Timestamp: s.TimestampMs,
			})
		}

		req.Timeseries = append(req.Timeseries, prompb.TimeSeries{
			Labels:  labels,
			Samples: samples,
		})
	}

	return req
}
