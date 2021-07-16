package examples

import (
	"context"
	"fmt"
	"github.com/BambooTuna/gooastream/builder"
	"github.com/BambooTuna/gooastream/stream"
	"sync"
	"time"
)

func ComplexConstructedStream() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	var wg sync.WaitGroup

	n := 10
	list := make([]interface{}, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		list[i] = i
	}
	graphBuilder := builder.NewGraphBuilder()
	source := graphBuilder.AddSource(stream.NewSliceSource(list))
	balance := graphBuilder.AddBalance(builder.NewBalance(3))
	merge := graphBuilder.AddMerge(builder.NewMerge(2))
	garbageSink := graphBuilder.AddSink(stream.NewSink(func(i interface{}) error {
		fmt.Println(i)
		wg.Done()
		return nil
	}))
	sink := graphBuilder.AddSink(stream.NewSink(func(i interface{}) error {
		fmt.Println(i)
		wg.Done()
		return nil
	}))

	/*
		source ~> balance ~> merge ~> sink
				  balance ~> merge
				  balance ~> garbageSink
	*/
	source.Out().Wire(balance.In())
	balance.Out()[0].Wire(merge.In()[0])
	balance.Out()[1].Wire(merge.In()[1])
	balance.Out()[2].Wire(garbageSink.In())
	merge.Out().Wire(sink.In())

	runnable := graphBuilder.ToRunnable()
	done, runningCancel := runnable.Run(ctx)
	go func() {
		wg.Wait()
		runningCancel()
	}()

	// blocking until runningCancel is called
	done()
}
