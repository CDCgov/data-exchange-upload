package postprocessing

import (
	"context"
)

type PostProcessor struct {
	UploadBaseDir string
	UploadDir     string
}

type Event struct {
	ID       string
	Manifest map[string]string
	Target   string
}

func Worker(ctx context.Context, c chan Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-c:
			if err := Deliver(ctx, e.ID, e.Manifest, e.Target); err != nil {
				// TODO Retry
				//go func(e Event) {
				//	c <- e
				//}(e)
			}
		}
	}
}
