package delivery

import (
	"context"
	"fmt"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/prometheus/client_golang/prometheus"
)

var SpeedHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "dex_server_delivery_speed_bytes_per_second",
	Help:    "File delivery speed distribution",
	Buckets: prometheus.ExponentialBuckets(10, 2.5, 20),
})

func ObserveSpeed(next func(context.Context, *event.FileReady) error) func(context.Context, *event.FileReady) error {
	return func(ctx context.Context, e *event.FileReady) error {
		// TODO middleware func for upload id logger
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
			// TODO log warn that couldn't get size and skipping speed measurement
		} else {
			dur := time.Since(start)
			if dur > 0 {
				speed := float64(size) / dur.Seconds()
				SpeedHistogram.Observe(speed)
			}
		}

		return nil
	}
}
