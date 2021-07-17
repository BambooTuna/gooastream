package fs

import (
	"context"
	"fmt"
	"github.com/BambooTuna/gooastream/stream"
	"time"
)

func ExampleNewFileSource() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	source := NewFileSource(&FileSourceConfig{
		FilePath:   "./test.txt",
		BufferSize: 1024,
	})
	sink := stream.NewSink(func(i interface{}) error {
		fmt.Println(string(i.([]byte)))
		return nil
	})
	runnable := source.To(sink)
	done, runnableCancel := runnable.Run(ctx)
	defer runnableCancel()
	done()

	// Output:
	// bambootuna
}

func ExampleNewFileSink() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	source := stream.NewSliceSource([]interface{}{
		[]byte("bambootuna"),
	})
	sink := NewFileSink(&FileSinkConfig{
		FilePath: "./test.txt",
	})
	runnable := source.To(sink)
	done, runnableCancel := runnable.Run(ctx)
	defer runnableCancel()
	done()

	// Output:
}
