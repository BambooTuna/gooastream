package nats

import (
	"context"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/nats-io/nats.go"
)

func NewNatsFlow(ctx context.Context, sourceConf *SourceConfig, sinkConf *SinkConfig, conn *nats.Conn) stream.Flow {
	sink := NewNatsSink(ctx, sinkConf, conn)
	source := NewNatsSource(ctx, sourceConf, conn)
	return stream.FlowFromSinkAndSource(sink, source)
}
