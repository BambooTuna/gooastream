package ws

import (
	"context"
	"fmt"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/gorilla/websocket"
	"time"
)

func ExampleNewWebSocketFlow() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ch, source := stream.NewChannelSource(10)
	go func() {
		for range time.Tick(time.Second) {
			err := ch.Push(&Message{
				Type:    websocket.TextMessage,
				Payload: []byte("payload"),
			})
			if err != nil {
				break
			}
		}
	}()
	sink := stream.NewSink(func(i interface{}) error {
		fmt.Println(string(i.(*Message).Payload))
		return nil
	})

	flow, err := NewWebSocketClientFlow(
		&SourceConfig{
			PongWait:       time.Second * 6,
			MaxMessageSize: 1024,
		},
		&SinkConfig{
			WriteWait:  time.Second,
			PingPeriod: time.Second * 5,
		},
		"ws://localhost:8080",
	)
	if err != nil {
		return
	}

	runnable := source.Via(flow).To(sink)
	done, runningCancel := runnable.Run(ctx)
	defer runningCancel()

	// blocking
	done()

	// TODO: run websocket server and check output
	// Output:
}
