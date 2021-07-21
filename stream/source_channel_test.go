package stream

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func ExampleNewChannelSource() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(3)

	channel, source := NewChannelSource(10)
	sink := NewSink(func(i interface{}) error {
		fmt.Println(i)
		wg.Done()
		return nil
	})

	err := channel.Push(1)
	if err != nil {
		return
	}
	err = channel.Push(2)
	if err != nil {
		return
	}
	err = channel.Push(3)
	if err != nil {
		return
	}

	runnable := source.To(sink)
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
