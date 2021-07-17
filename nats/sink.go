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

func NewNatsSink(conf *SinkConfig, conn *nats.Conn) stream.Sink {
	in := queue.NewQueueEmpty(conf.Buffer)
	graphTree := stream.EmptyGraph()
	graphTree.AddWire(newNatsSinkWire(in, conf, conn))
	return stream.BuildSink(in, graphTree)
}

type natsSinkWire struct {
	from queue.OutQueue

	conf *SinkConfig
	conn *nats.Conn
}

func newNatsSinkWire(from queue.OutQueue, conf *SinkConfig, conn *nats.Conn) stream.Wire {
	return &natsSinkWire{
		from: from,
		conf: conf,
		conn: conn,
	}
}

func (a natsSinkWire) Run(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		cancel()
		a.from.Close()
	}()
T:
	for {
		select {
		case <-ctx.Done():
			break T
		default:
			data, err := a.from.Pop(ctx)
			if err != nil {
				break T
			}
			switch msg := data.(type) {
			case *nats.Msg:
				err = a.conn.PublishMsg(msg)
				if err != nil {
					break T
				}
			case []byte:
				err = a.conn.Publish(a.conf.ByteSubject, msg)
				if err != nil {
					break T
				}
			default:
				continue
			}
		}
	}
}

var _ stream.Wire = (*natsSinkWire)(nil)
