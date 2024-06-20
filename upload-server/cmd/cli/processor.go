package cli

import (
	"context"
	"runtime"
	"sync"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
)

func StartProcessorWorkers(ctx context.Context, c chan postprocessing.Event) {
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			postprocessing.Worker(ctx, c)
			wg.Done()
		}()
	}

	wg.Wait()
}
