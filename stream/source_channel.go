package stream

import (
	"context"
	"fmt"
	"github.com/BambooTuna/gooastream/queue"
	"sync"
)

/*
	SourceChannel
	This is input channel.
	If this is closed, it will be transmitted to the entire stream.
*/
type SourceChannel struct {
	mu       sync.RWMutex
	ch       chan interface{}
	isClosed bool
}

func (a *SourceChannel) Push(in interface{}) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.isClosed {
		return fmt.Errorf("source chan is closed")
	}
	select {
	case a.ch <- in:
		return nil
	default:
		return fmt.Errorf("source chan is clogged")
	}
}
func (a *SourceChannel) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.isClosed {
		close(a.ch)
	}
	a.isClosed = true
}

/*
	NewChannelSource
	Create SourceChannel and Source with a buffer.
	Have no input port and one output port.
	If the downstream is clogged, what you send to SourceChannel will be stored in the buffer.
*/
func NewChannelSource(chanBuffer int, options ...queue.Option) (*SourceChannel, Source) {
	in := SourceChannel{
		ch: make(chan interface{}, chanBuffer),
	}
	out := queue.NewQueueEmpty(options...)
	graphTree := EmptyGraph()
	graphTree.AddWire(newChanSourceWire(out, &in))
	return &in, BuildSource(out, graphTree)
}

type chanSourceWire struct {
	to     queue.InQueue
	inChan *SourceChannel
}

func newChanSourceWire(to queue.InQueue, inChan *SourceChannel) Wire {
	return &chanSourceWire{
		to:     to,
		inChan: inChan,
	}
}

func (a chanSourceWire) Run(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		cancel()
		a.inChan.Close()
		a.to.Close()
	}()
T:
	for {
		select {
		case <-ctx.Done():
			break T
		case v, ok := <-a.inChan.ch:
			if !ok {
				break T
			}
			err := a.to.Push(ctx, v)
			if err != nil {
				break T
			}
		}
	}
}

var _ Wire = (*chanSourceWire)(nil)
