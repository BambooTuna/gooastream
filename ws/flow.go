package ws

import (
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/gorilla/websocket"
	"time"
)

type FlowConfig struct {
	PongWait       time.Duration
	MaxMessageSize int64
}

func NewWebSocketFlow(sourceConf *SourceConfig, sinkConf *SinkConfig, conn *websocket.Conn, options ...queue.Option) stream.Flow {
	source := NewWebSocketSource(sourceConf, conn, options...)
	sink := NewWebSocketSink(sinkConf, conn, options...)
	return stream.FlowFromSinkAndSource(sink, source)
}

func NewWebSocketClientFlow(sourceConf *SourceConfig, sinkConf *SinkConfig, url string, options ...queue.Option) (stream.Flow, error) {
	return NewWebSocketClientFlowWithDialer(sourceConf, sinkConf, url, websocket.DefaultDialer, options...)
}

func NewWebSocketClientFlowWithDialer(sourceConf *SourceConfig, sinkConf *SinkConfig, url string, dialer *websocket.Dialer, options ...queue.Option) (stream.Flow, error) {
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return NewWebSocketFlow(sourceConf, sinkConf, conn, options...), nil
}
