package stream

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func ExampleNewFlattenFlow() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(3)

	source := NewSliceSource([]interface{}{
		[]interface{}{1, 2, 3},
	})

	flow := NewFlattenFlow()
	sink := NewSink(func(i interface{}) error {
		v := i.(int)
		fmt.Println(v)
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
	// 1
	// 2
	// 3
}
