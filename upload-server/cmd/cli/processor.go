package cli

import (
	"context"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	"sync"
)

func MakeEventListener(appConfig appconfig.AppConfig, c chan event.FileReadyEvent) postprocessing.EventProcessable {
	return &postprocessing.MemoryEventListener{
		C: c,
	}
}

// Maybe separate creating of workers from starting of them.  Then can just pass in workers here.
func StartEventListener(ctx context.Context, listener postprocessing.EventProcessable) {
	for {
		var wg sync.WaitGroup
		events := listener.GetEventBatch(5)
		select {
		case <-ctx.Done():
			return
		default:
			for _, e := range events {
				wg.Add(1)
				e := e // TODO is this necessary?
				go func() {
					defer wg.Done()
					listener.Process(ctx, e)
				}()
			}
			wg.Wait()
		}
	}

	//var wg sync.WaitGroup
	//numWorkers := len(workers)
	//wg.Add(numWorkers)
	//
	//for i := 0; i < numWorkers; i++ {
	//	i := i // TODO is this needed?
	//	go func() {
	//		//postprocessing.Worker(ctx, c)
	//		workers[i].DoWork(ctx)
	//		wg.Done()
	//	}()
	//}
	//
	//wg.Wait()
}
