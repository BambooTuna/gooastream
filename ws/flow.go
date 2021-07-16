package ws

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/gorilla/websocket"
	"time"
)

type FlowConfig struct {
	PongWait       time.Duration
	MaxMessageSize int64
	Buffer         int
}

type webSocketFlow struct {
	in        queue.Queue
	out       queue.Queue
	graphTree *stream.GraphTree

	conf *FlowConfig
	conn *websocket.Conn
}

var _ stream.Flow = (*webSocketFlow)(nil)

func NewWebSocketFlow(ctx context.Context, sourceConf *SourceConfig, sinkConf *SinkConfig, conn *websocket.Conn) stream.Flow {
	source := NewWebSocketSource(ctx, sourceConf, conn)
	sink := NewWebSocketSink(ctx, sinkConf, conn)
	return stream.FlowFromSinkAndSource(sink, source)
}

func NewWebSocketClientFlow(ctx context.Context, sourceConf *SourceConfig, sinkConf *SinkConfig, url string) (stream.Flow, error) {
	return NewWebSocketClientFlowWithDialer(ctx, sourceConf, sinkConf, url, websocket.DefaultDialer)
}

func NewWebSocketClientFlowWithDialer(ctx context.Context, sourceConf *SourceConfig, sinkConf *SinkConfig, url string, dialer *websocket.Dialer) (stream.Flow, error) {
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return NewWebSocketFlow(ctx, sourceConf, sinkConf, conn), nil
}

func (a webSocketFlow) Via(flow stream.Flow) stream.Flow {
	return stream.BuildFlow(
		a.In(),
		flow.Out(),
		a.GraphTree().
			Append(stream.PassThrowGraph(a.Out(), flow.In())).
			Append(flow.GraphTree()),
	)
}

func (a webSocketFlow) To(sink stream.Sink) stream.Sink {
	return stream.BuildSink(
		a.In(),
		a.GraphTree().
			Append(stream.PassThrowGraph(a.Out(), sink.In())).
			Append(sink.GraphTree()),
	)
}

func (a webSocketFlow) In() queue.Queue {
	return a.in
}

func (a webSocketFlow) Out() queue.Queue {
	return a.out
}

func (a webSocketFlow) GraphTree() *stream.GraphTree {
	return a.graphTree
}

func (a webSocketFlow) connect(ctx context.Context) {

}
