package ws

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/gorilla/websocket"
	"time"
)

type SourceConfig struct {
	PongWait       time.Duration
	MaxMessageSize int64
}

func NewWebSocketSource(conf *SourceConfig, conn *websocket.Conn, options ...queue.Option) stream.Source {
	out := queue.NewQueueEmpty(options...)
	graphTree := stream.EmptyGraph()
	graphTree.AddWire(newWebSocketSourceWire(out, conf, conn))
	return stream.BuildSource(out, graphTree)
}

func NewWebSocketClientSource(conf *SourceConfig, url string, options ...queue.Option) (stream.Source, error) {
	return NewWebSocketClientSourceWithDialer(conf, url, websocket.DefaultDialer, options...)
}

func NewWebSocketClientSourceWithDialer(conf *SourceConfig, url string, dialer *websocket.Dialer, options ...queue.Option) (stream.Source, error) {
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return NewWebSocketSource(conf, conn, options...), nil
}

type webSocketSourceWire struct {
	to queue.InQueue

	conf *SourceConfig
	conn *websocket.Conn
}

func newWebSocketSourceWire(to queue.InQueue, conf *SourceConfig, conn *websocket.Conn) stream.Wire {
	return &webSocketSourceWire{
		to:   to,
		conf: conf,
		conn: conn,
	}
}

func (a webSocketSourceWire) Run(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		cancel()
		_ = a.conn.Close()
		a.to.Close()
	}()
	a.conn.SetReadLimit(a.conf.MaxMessageSize)
	_ = a.conn.SetReadDeadline(time.Now().Add(a.conf.PongWait))
	a.conn.SetPongHandler(func(string) error { _ = a.conn.SetReadDeadline(time.Now().Add(a.conf.PongWait)); return nil })
	for {
		t, b, err := a.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			}
			break
		}
		err = a.to.Push(ctx, &Message{
			Type:    t,
			Payload: b,
		})
		if err != nil {
			break
		}
	}
}

var _ stream.Wire = (*webSocketSourceWire)(nil)
