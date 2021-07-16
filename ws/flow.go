package ws

import (
	"context"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/gorilla/websocket"
	"time"
)

type FlowConfig struct {
	PongWait       time.Duration
	MaxMessageSize int64
	Buffer         int
}

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
