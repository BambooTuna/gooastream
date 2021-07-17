package nats

import (
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/nats-io/nats.go"
)

func NewNatsFlow(sourceConf *SourceConfig, sinkConf *SinkConfig, conn *nats.Conn, options ...queue.Option) stream.Flow {
	sink := NewNatsSink(sinkConf, conn, options...)
	source := NewNatsSource(sourceConf, conn, options...)
	return stream.FlowFromSinkAndSource(sink, source)
}
