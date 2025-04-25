package delivery

import (
	"context"
	"fmt"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/prometheus/client_golang/prometheus"
)

var SpeedHistograms = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "dex_server_delivery_speed_bytes_per_second",
	Help:    "File delivery speed distribution",
	Buckets: prometheus.ExponentialBuckets(10, 2.5, 20),
}, []string{"target"})

func ObserveSpeed(next func(context.Context, *event.FileReady) error) func(context.Context, *event.FileReady) error {
	return func(ctx context.Context, e *event.FileReady) error {
		logger := sloger.FromContext(ctx)

		src, ok := GetSource(UploadSrc)
		if !ok {
			return fmt.Errorf("failed to get source for file delivery %+v", e)
		}

		start := time.Now()
		err := next(ctx, e)
		// don't take a measurement if the delivery failed
		if err != nil {
			return err
		}

		size, err := src.GetSize(ctx, e.UploadId)
		if err != nil {
			logger.Warn("skipping delivery speed measurement due to error getting file size", "error", err)
		} else {
			dur := time.Since(start)
			if dur > 0 {
				speed := float64(size) / dur.Seconds()
				SpeedHistograms.With(prometheus.Labels{"target": e.DestinationTarget}).Observe(speed)
			}
		}

		return nil
	}
}
