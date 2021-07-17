package examples

import (
	"context"
	"fmt"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/BambooTuna/gooastream/ws"
	"github.com/gorilla/websocket"
	"time"
)

func WebsocketStream() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ch, source := stream.NewChannelSource(0)
	go func() {
		for range time.Tick(time.Second) {
			err := ch.Push(ctx, &ws.Message{
				Type:    websocket.TextMessage,
				Payload: []byte("payload"),
			})
			if err != nil {
				break
			}
		}
	}()
	sink := stream.NewSink(func(i interface{}) error {
		fmt.Println(string(i.(*ws.Message).Payload))
		return nil
	})

	flow, err := ws.NewWebSocketClientFlow(
		&ws.SourceConfig{
			PongWait:       time.Second * 6,
			MaxMessageSize: 1024,
			Buffer:         0,
		},
		&ws.SinkConfig{
			WriteWait:  time.Second,
			PingPeriod: time.Second * 5,
			Buffer:     0,
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
}
