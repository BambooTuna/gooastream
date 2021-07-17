package ws

import (
	"github.com/BambooTuna/gooastream/stream"
	"github.com/gorilla/websocket"
	"time"
)

type FlowConfig struct {
	PongWait       time.Duration
	MaxMessageSize int64
	Buffer         int
}

func NewWebSocketFlow(sourceConf *SourceConfig, sinkConf *SinkConfig, conn *websocket.Conn) stream.Flow {
	source := NewWebSocketSource(sourceConf, conn)
	sink := NewWebSocketSink(sinkConf, conn)
	return stream.FlowFromSinkAndSource(sink, source)
}

func NewWebSocketClientFlow(sourceConf *SourceConfig, sinkConf *SinkConfig, url string) (stream.Flow, error) {
	return NewWebSocketClientFlowWithDialer(sourceConf, sinkConf, url, websocket.DefaultDialer)
}

func NewWebSocketClientFlowWithDialer(sourceConf *SourceConfig, sinkConf *SinkConfig, url string, dialer *websocket.Dialer) (stream.Flow, error) {
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return NewWebSocketFlow(sourceConf, sinkConf, conn), nil
}
