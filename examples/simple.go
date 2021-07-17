package examples

import (
	"context"
	"fmt"
	"github.com/BambooTuna/gooastream/stream"
	"sync"
	"time"
)

func SimpleRunnableStream() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var wg sync.WaitGroup

	n := 5
	list := make([]interface{}, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		list[i] = i
	}
	source := stream.NewSliceSource(list)
	flow := stream.NewBufferFlow()
	sink := stream.NewSink(func(i interface{}) error {
		fmt.Println(i)
		wg.Done()
		return nil
	})

	runnable := source.Via(flow).To(sink)
	done, runningCancel := runnable.Run(ctx)
	go func() {
		wg.Wait()
		runningCancel()
	}()

	// blocking until runningCancel is called
	done()
}
