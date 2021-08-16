package stream

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
	"reflect"
)

type flattenFlowWire struct {
	from queue.OutQueue
	to   queue.InQueue
}

func NewFlattenFlow(options ...queue.Option) Flow {
	in := queue.NewQueueEmpty(options...)
	out := queue.NewQueueEmpty(options...)
	graphTree := EmptyGraph()
	graphTree.AddWire(newFlattenFlowWire(in, out))
	return BuildFlow(in, out, graphTree)
}

func newFlattenFlowWire(from queue.OutQueue, to queue.InQueue) Wire {
	return &flattenFlowWire{
		from: from,
		to:   to,
	}
}

func (a flattenFlowWire) Run(ctx context.Context, cancel context.CancelFunc) {
	var err error
	defer func() {
		cancel()
		a.from.Close()
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
		default:
			v, err := a.from.Pop(ctx)
			if err != nil {
				break T
			}
			switch reflect.TypeOf(v).Kind() {
			case reflect.Slice:
				s := reflect.ValueOf(v)
				for i := 0; i < s.Len(); i++ {
					err = a.to.Push(ctx, s.Index(i).Interface())
					if err != nil {
						break T
					}
				}
			case reflect.Array:
				s := reflect.ValueOf(v)
				for i := 0; i < s.Len(); i++ {
					err = a.to.Push(ctx, s.Index(i).Interface())
					if err != nil {
						break T
					}
				}
			default:
				err = a.to.Push(ctx, v)
				if err != nil {
					break T
				}
			}
		}
	}
}

var _ Wire = (*flattenFlowWire)(nil)
