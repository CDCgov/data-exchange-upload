package cli

import (
	"context"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"runtime"
	"sync"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
)

func StartProcessorWorkers(ctx context.Context, c chan event.FileReadyEvent) {
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
