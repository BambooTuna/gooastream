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

type natsSource struct {
	out       queue.Queue
	graphTree *stream.GraphTree

	conf *SourceConfig
	conn *nats.Conn
}

func NewNatsSource(ctx context.Context, conf *SourceConfig, conn *nats.Conn) stream.Source {
	out := queue.NewQueueEmpty(conf.Buffer)
	source := natsSource{
		out:       out,
		graphTree: stream.EmptyGraph(),

		conf: conf,
		conn: conn,
	}
	go source.connect(ctx)
	return &source
}

func (a natsSource) Via(flow stream.Flow) stream.Source {
	return stream.BuildSource(
		flow.Out(),
		a.GraphTree().
			Append(stream.PassThrowGraph(a.Out(), flow.In())).
			Append(flow.GraphTree()),
	)
}

func (a natsSource) To(sink stream.Sink) stream.Runnable {
	return stream.NewRunnable(
		a.GraphTree().
			Append(stream.PassThrowGraph(a.Out(), sink.In())).
			Append(sink.GraphTree()),
	)
}

func (a natsSource) Out() queue.Queue {
	return a.out
}

func (a natsSource) GraphTree() *stream.GraphTree {
	return a.graphTree
}

func (a natsSource) connect(ctx context.Context) {
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
					err := a.out.Push(ctx, msg)
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

var _ stream.Source = (*natsSource)(nil)
