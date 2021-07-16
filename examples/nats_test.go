package examples

import (
	"context"
	"fmt"
	ns "github.com/BambooTuna/gooastream/nats"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/nats-io/nats.go"
	"sync"
	"time"
)

func NatsStream() {
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
	}, 0)
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
	flow := ns.NewNatsFlow(
		ctx,
		&ns.SourceConfig{
			Subjects: []string{"subject1", "subject2"},
			Buffer:   0,
		},
		&ns.SinkConfig{
			ByteSubject: "default",
			Buffer:      0,
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
	// subject1 data1
	// subject1 data2
	// subject2 data3
}
