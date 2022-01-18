package builder

import (
	"context"
	"fmt"
	"github.com/BambooTuna/gooastream/stream"
	"sync"
	"time"
)

/*
	You can create a stream.Runnable by calling GraphBuilder.ToRunnable after building a closed graph.
*/
func ExampleGraphBuilder_ToRunnable() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	var wg sync.WaitGroup

	n := 10
	list := make([]interface{}, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		list[i] = i
	}
	graphBuilder := NewGraphBuilder()
	source := graphBuilder.AddSource(stream.NewSliceSource(list))
	balance := graphBuilder.AddBalance(NewBalance(3))
	merge := graphBuilder.AddMerge(NewMerge(2))
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

	source.Out().Wire(merge.In()[0])
	merge.Out().Wire(balance.In())
	balance.Out()[0].Wire(sink.In())
	balance.Out()[1].Wire(merge.In()[1])

	runnable := graphBuilder.ToRunnable()

	done, runningCancel := runnable.Run(ctx)
	go func() {
		wg.Wait()
		runningCancel()
	}()

	// blocking until runningCancel is called
	done()

	// Unordered output:
	// 1
	// 6
	// 9
	// 0
	// 2
	// 3
	// 4
	// 7
	// 5
	// 8
}

/*
	You can create a stream.Source by calling GraphBuilder.ToSource after building a not closed graph ( has no input and one output graph ).
*/
func ExampleGraphBuilder_ToSource() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	var wg sync.WaitGroup

	n := 10
	list := make([]interface{}, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		list[i] = i
	}
	graphBuilder := NewGraphBuilder()
	source := graphBuilder.AddSource(stream.NewSliceSource(list))
	balance := graphBuilder.AddBalance(NewBalance(3))
	merge := graphBuilder.AddMerge(NewMerge(2))
	garbage := graphBuilder.AddSink(stream.NewSink(func(i interface{}) error {
		fmt.Println(i)
		wg.Done()
		return nil
	}))

	/*
		source ~> balance ~> merge ~>
				  balance ~> merge
				  balance ~> garbage
	*/
	source.Out().Wire(balance.In())
	balance.Out()[0].Wire(merge.In()[0])
	balance.Out()[1].Wire(merge.In()[1])
	balance.Out()[2].Wire(garbage.In())

	buildSource := graphBuilder.ToSource(merge.Out())
	sink := stream.NewSink(func(i interface{}) error {
		fmt.Println(i)
		wg.Done()
		return nil
	})

	runnable := buildSource.To(sink)
	done, runningCancel := runnable.Run(ctx)
	go func() {
		wg.Wait()
		runningCancel()
	}()

	// blocking until runningCancel is called
	done()

	// Unordered output:
	// 1
	// 6
	// 9
	// 0
	// 2
	// 3
	// 4
	// 7
	// 5
	// 8
}

func ExampleNewBroadcast() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	var wg sync.WaitGroup

	n := 5
	list := make([]interface{}, n)
	wg.Add(n * 3)
	for i := 0; i < n; i++ {
		list[i] = i
	}
	graphBuilder := NewGraphBuilder()
	source := graphBuilder.AddSource(stream.NewSliceSource(list))
	balance := graphBuilder.AddBalance(NewBroadcast(3))
	merge := graphBuilder.AddMerge(NewMerge(2))
	garbage := graphBuilder.AddSink(stream.NewSink(func(i interface{}) error {
		fmt.Println(i)
		wg.Done()
		return nil
	}))

	/*
		source ~> broadcast ~> merge ~>
				  broadcast ~> merge
				  broadcast ~> garbage
	*/
	source.Out().Wire(balance.In())
	balance.Out()[0].Wire(merge.In()[0])
	balance.Out()[1].Wire(merge.In()[1])
	balance.Out()[2].Wire(garbage.In())

	buildSource := graphBuilder.ToSource(merge.Out())
	sink := stream.NewSink(func(i interface{}) error {
		fmt.Println(i)
		wg.Done()
		return nil
	})

	runnable := buildSource.To(sink)
	done, runningCancel := runnable.Run(ctx)
	go func() {
		wg.Wait()
		runningCancel()
	}()

	// blocking until runningCancel is called
	done()

	// Unordered output:
	// 0
	// 0
	// 0
	// 1
	// 1
	// 1
	// 2
	// 2
	// 2
	// 3
	// 3
	// 3
	// 4
	// 4
	// 4
}
