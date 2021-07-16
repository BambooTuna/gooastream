package stream

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func ExampleRunnable_Run() {
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
	flow := NewBufferFlow(0)
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
