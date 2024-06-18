package cli

import (
	"context"
	"runtime"
	"sync"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
)

func StartProcessorWorkers(ctx context.Context) (chan postprocessing.Event, *sync.WaitGroup) {
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()
	c := make(chan postprocessing.Event, numWorkers)
	for i := 0; i < numWorkers; i++ {
		go postprocessing.Worker(ctx, c, &wg)
	}
	//TODO could do something like return sa waitgroup for elegant shutdown
	return c, &wg
}
