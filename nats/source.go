package nats

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/nats-io/nats.go"
	"sync"
)

type SourceConfig struct {
	Subjects []string
	Buffer   int
}

func NewNatsSource(conf *SourceConfig, conn *nats.Conn) stream.Source {
	out := queue.NewQueueEmpty(conf.Buffer)
	graphTree := stream.EmptyGraph()
	graphTree.AddWire(newNatsSourceWire(out, conf, conn))
	return stream.BuildSource(out, graphTree)
}

type natsSourceWire struct {
	to queue.InQueue

	conf *SourceConfig
	conn *nats.Conn
}

func newNatsSourceWire(to queue.InQueue, conf *SourceConfig, conn *nats.Conn) stream.Wire {
	return natsSourceWire{
		to:   to,
		conf: conf,
		conn: conn,
	}
}

func (a natsSourceWire) Run(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		cancel()
		a.to.Close()
	}()
	var wg sync.WaitGroup
	for _, subject := range a.conf.Subjects {
		wg.Add(1)
		go func(subject string) {
			ch := make(chan *nats.Msg)
			subscribe, err := a.conn.ChanSubscribe(subject, ch)
			if err != nil {
				return
			}
			defer func() {
				_ = subscribe.Unsubscribe()
				close(ch)
			}()
		T:
			for {
				select {
				case <-ctx.Done():
					break T
				case msg, ok := <-ch:
					if !ok {
						break T
					}
					err := a.to.Push(ctx, msg)
					if err != nil {
						break T
					}
				}
			}
			wg.Done()
		}(subject)
	}
	wg.Wait()
}

var _ stream.Wire = (*natsSourceWire)(nil)
