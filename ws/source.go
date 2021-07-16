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
	Buffer         int
}

type webSocketSource struct {
	out       queue.Queue
	graphTree *stream.GraphTree

	conf *SourceConfig
	conn *websocket.Conn
}

var _ stream.Source = (*webSocketSource)(nil)

func NewWebSocketSource(ctx context.Context, conf *SourceConfig, conn *websocket.Conn) stream.Source {
	out := queue.NewQueueEmpty(conf.Buffer)
	source := webSocketSource{
		out:       out,
		graphTree: stream.EmptyGraph(),

		conf: conf,
		conn: conn,
	}
	go source.connect(ctx)
	return &source
}

func NewWebSocketClientSource(ctx context.Context, conf *SourceConfig, url string) (stream.Source, error) {
	return NewWebSocketClientSourceWithDialer(ctx, conf, url, websocket.DefaultDialer)
}

func NewWebSocketClientSourceWithDialer(ctx context.Context, conf *SourceConfig, url string, dialer *websocket.Dialer) (stream.Source, error) {
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return NewWebSocketSource(ctx, conf, conn), nil
}

func (a webSocketSource) Via(flow stream.Flow) stream.Source {
	return stream.BuildSource(
		flow.Out(),
		a.GraphTree().
			Append(stream.PassThrowGraph(a.Out(), flow.In())).
			Append(flow.GraphTree()),
	)
}

func (a webSocketSource) To(sink stream.Sink) stream.Runnable {
	return stream.NewRunnable(
		a.GraphTree().
			Append(stream.PassThrowGraph(a.Out(), sink.In())).
			Append(sink.GraphTree()),
	)
}

func (a webSocketSource) Out() queue.Queue {
	return a.out
}

func (a webSocketSource) GraphTree() *stream.GraphTree {
	return a.graphTree
}

func (a webSocketSource) connect(ctx context.Context) {
	defer func() {
		_ = a.conn.Close()
		a.out.Close()
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
		err = a.out.Push(ctx, &Message{
			Type:    t,
			Payload: b,
		})
		if err != nil {
			break
		}
	}
}
