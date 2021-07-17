package nats

import (
	"github.com/BambooTuna/gooastream/stream"
	"github.com/nats-io/nats.go"
)

func NewNatsFlow(sourceConf *SourceConfig, sinkConf *SinkConfig, conn *nats.Conn) stream.Flow {
	sink := NewNatsSink(sinkConf, conn)
	source := NewNatsSource(sourceConf, conn)
	return stream.FlowFromSinkAndSource(sink, source)
}
