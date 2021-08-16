package stream

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
	"time"
)

type TickerSourceConfig struct {
	Duration time.Duration
	Element  interface{}
}

type tickerSourceWire struct {
	to queue.InQueue

	duration time.Duration
	element  interface{}
}

func NewTickerSource(conf *TickerSourceConfig, options ...queue.Option) Source {
	out := queue.NewQueueEmpty(options...)
	graphTree := EmptyGraph()
	graphTree.AddWire(newTickerSourceWire(out, conf.Duration, conf.Element))
	return BuildSource(out, graphTree)
}

func newTickerSourceWire(to queue.InQueue, duration time.Duration, element interface{}) Wire {
	return &tickerSourceWire{
		to:       to,
		duration: duration,
		element:  element,
	}
}

func (a tickerSourceWire) Run(ctx context.Context, cancel context.CancelFunc) {
	ticker := time.NewTicker(a.duration)
	var err error
	defer func() {
		ticker.Stop()
		cancel()
		a.to.Close()
		if err != nil {
			Log().Errorf("%v", err)
		}
	}()
T:
	for {
		select {
		case <-ctx.Done():
			break T
		case <-ticker.C:
			err = a.to.Push(ctx, a.element)
			if err != nil {
				break T
			}
		}
	}
}

var _ Wire = (*tickerSourceWire)(nil)
