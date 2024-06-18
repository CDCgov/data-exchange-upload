package postprocessing

import (
	"context"
	"sync"
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

func Worker(ctx context.Context, c chan Event, wg *sync.WaitGroup) {
	for e := range c {
		if err := Deliver(ctx, e.ID, e.Manifest, e.Target); err != nil {
			go func(e Event) {
				c <- e
			}(e)
		}
		wg.Done()
	}
}
