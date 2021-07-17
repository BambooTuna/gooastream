package stream

import (
	"context"
	"fmt"
	"github.com/BambooTuna/gooastream/queue"
	"sync"
	"time"
)

func ExampleRunnable_Run() {
	queue.SetGlobalBuffer(100)
	queue.SetGlobalStrategy(queue.Fail)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var wg sync.WaitGroup

	n := 5
	list := make([]interface{}, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		list[i] = i
	}
	source := NewSliceSource(list)
	flow := NewBufferFlow()
	sink := NewSink(func(i interface{}) error {
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

	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
}
