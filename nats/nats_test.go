package nats

import (
	"context"
	"fmt"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/nats-io/nats.go"
	"sync"
	"time"
)

func ExampleNewNatsFlow() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(3)
	source := stream.NewSliceSource([]interface{}{
		&nats.Msg{
			Subject: "subject1",
			Data:    []byte("data1"),
		},
		&nats.Msg{
			Subject: "subject1",
			Data:    []byte("data2"),
		},
		&nats.Msg{
			Subject: "subject2",
			Data:    []byte("data3"),
		},
	})
	delayFlow := stream.NewMapFlow(func(i interface{}) (interface{}, error) {
		time.Sleep(time.Millisecond * 100)
		return i, nil
	})
	sink := stream.NewSink(func(i interface{}) error {
		msg := i.(*nats.Msg)
		fmt.Println(msg.Subject, string(msg.Data))
		wg.Done()
		return nil
	})

	conn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return
	}
	flow := NewNatsFlow(
		&SourceConfig{
			Subjects: []string{"subject1", "subject2"},
		},
		&SinkConfig{
			ByteSubject: "default",
		},
		conn,
	)

	runnable := source.Via(delayFlow).Via(flow).To(sink)
	done, runningCancel := runnable.Run(ctx)
	defer runningCancel()

	go func() {
		wg.Wait()
		runningCancel()
	}()
	// blocking
	done()

	// Output:

	// TODO: run nats server and check output
	// subject1 data1
	// subject1 data2
	// subject2 data3
}
