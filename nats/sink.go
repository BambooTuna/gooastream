package nats

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/nats-io/nats.go"
)

type SinkConfig struct {
	// if data type is not *nats.Msg but []byte, it will send to ByteSubject
	ByteSubject string

	Buffer int
}

type natsSink struct {
	in        queue.Queue
	graphTree *stream.GraphTree

	conf *SinkConfig
	conn *nats.Conn
}

func NewNatsSink(ctx context.Context, conf *SinkConfig, conn *nats.Conn) stream.Sink {
	in := queue.NewQueueEmpty(conf.Buffer)
	sink := natsSink{
		in:        in,
		graphTree: stream.EmptyGraph(),

		conf: conf,
		conn: conn,
	}
	go sink.connect(ctx)
	return &sink
}

func (a natsSink) Dummy() {
}

func (a natsSink) In() queue.Queue {
	return a.in
}

func (a natsSink) GraphTree() *stream.GraphTree {
	return a.graphTree
}

func (a natsSink) connect(ctx context.Context) {
	for {
		data, err := a.in.Pop(ctx)
		if err != nil {
			break
		}
		switch msg := data.(type) {
		case *nats.Msg:
			err = a.conn.PublishMsg(msg)
			if err != nil {
				break
			}
		case []byte:
			err = a.conn.Publish(a.conf.ByteSubject, msg)
			if err != nil {
				break
			}
		default:
			continue
		}
	}
}

var _ stream.Sink = (*natsSink)(nil)
