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

func NewWebSocketSink(conf *SinkConfig, conn *websocket.Conn) stream.Sink {
	in := queue.NewQueueEmpty(conf.Buffer)
	graphTree := stream.EmptyGraph()
	graphTree.AddWire(newWebSocketSinkWire(in, conf, conn))
	return stream.BuildSink(in, graphTree)
}

func NewWebSocketClientSink(conf *SinkConfig, url string) (stream.Sink, error) {
	return NewWebSocketClientSinkWithDialer(conf, url, websocket.DefaultDialer)
}

func NewWebSocketClientSinkWithDialer(conf *SinkConfig, url string, dialer *websocket.Dialer) (stream.Sink, error) {
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return NewWebSocketSink(conf, conn), nil
}

type webSocketSinkWire struct {
	from queue.OutQueue

	conf *SinkConfig
	conn *websocket.Conn
}

func newWebSocketSinkWire(from queue.OutQueue, conf *SinkConfig, conn *websocket.Conn) stream.Wire {
	return &webSocketSinkWire{
		from: from,
		conf: conf,
		conn: conn,
	}
}

func (a webSocketSinkWire) Run(ctx context.Context, cancel context.CancelFunc) {
	ticker := time.NewTicker(a.conf.PingPeriod)
	ch := make(chan interface{}, 0)
	defer func() {
		cancel()
		ticker.Stop()
		a.from.Close()
		_ = a.conn.Close()
	}()

	go func() {
		for {
			data, err := a.from.Pop(ctx)
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
		case <-ctx.Done():
			break T
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

var _ stream.Wire = (*webSocketSinkWire)(nil)
