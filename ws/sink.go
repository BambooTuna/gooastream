package ws

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/gorilla/websocket"
	"time"
)

type SinkConfig struct {
	WriteWait  time.Duration
	PingPeriod time.Duration
	Buffer     int
}

type webSocketSink struct {
	in        queue.Queue
	graphTree *stream.GraphTree

	conf *SinkConfig
	conn *websocket.Conn
}

var _ stream.Sink = (*webSocketSink)(nil)

func NewWebSocketSink(ctx context.Context, conf *SinkConfig, conn *websocket.Conn) stream.Sink {
	in := queue.NewQueueEmpty(conf.Buffer)
	sink := webSocketSink{
		in:        in,
		graphTree: stream.EmptyGraph(),

		conf: conf,
		conn: conn,
	}
	go sink.connect(ctx)
	return &sink
}

func NewWebSocketClientSink(ctx context.Context, conf *SinkConfig, url string) (stream.Sink, error) {
	return NewWebSocketClientSinkWithDialer(ctx, conf, url, websocket.DefaultDialer)
}

func NewWebSocketClientSinkWithDialer(ctx context.Context, conf *SinkConfig, url string, dialer *websocket.Dialer) (stream.Sink, error) {
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return NewWebSocketSink(ctx, conf, conn), nil
}

func (a webSocketSink) Dummy() {
}

func (a webSocketSink) In() queue.Queue {
	return a.in
}

func (a webSocketSink) GraphTree() *stream.GraphTree {
	return a.graphTree
}

func (a webSocketSink) connect(ctx context.Context) {
	ticker := time.NewTicker(a.conf.PingPeriod)
	ch := make(chan interface{}, 0)
	defer func() {
		ticker.Stop()
		a.in.Close()
		_ = a.conn.Close()
	}()

	go func() {
		for {
			data, err := a.in.Pop(ctx)
			if err != nil {
				break
			}
			ch <- data
		}
		close(ch)
	}()

T:
	for {
		select {
		case data, ok := <-ch:
			if !ok {
				break T
			}
			_ = a.conn.SetWriteDeadline(time.Now().Add(a.conf.WriteWait))
			message, ok := data.(*Message)
			if !ok {
				continue
			}
			err := a.conn.WriteMessage(message.Type, message.Payload)
			if err != nil {
				break T
			}
		case <-ticker.C:
			_ = a.conn.SetWriteDeadline(time.Now().Add(a.conf.WriteWait))
			if err := a.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				break T
			}
		}
	}
}
