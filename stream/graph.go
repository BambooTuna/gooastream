package stream

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
)

type (
	Mat       uint8
	GraphTree struct {
		wires []*wire
	}
	wire struct {
		from queue.OutQueue
		to   queue.InQueue
		task func(interface{}) (interface{}, error)
	}
)

const (
	MatNone Mat = iota
	MatLeft
	MatRight
	MatBoth
)

var emptyTaskFunc = func(i interface{}) (interface{}, error) {
	return i, nil
}

func EmptyGraph() *GraphTree {
	return &GraphTree{
		wires: []*wire{},
	}
}

func PassThrowGraph(from queue.OutQueue, to queue.InQueue) *GraphTree {
	return &GraphTree{
		wires: []*wire{
			{
				from: from,
				to:   to,
				task: emptyTaskFunc,
			},
		},
	}
}

func MapGraph(from queue.OutQueue, to queue.InQueue, f func(interface{}) (interface{}, error)) *GraphTree {
	return &GraphTree{
		wires: []*wire{
			{
				from: from,
				to:   to,
				task: f,
			},
		},
	}
}

func (a *GraphTree) Append(child *GraphTree) *GraphTree {
	return &GraphTree{wires: append(a.wires, child.wires...)}
}

func (a *GraphTree) Add(child *GraphTree) {
	a.wires = append(a.wires, child.wires...)
}

// non blocking
func (a *GraphTree) Run(ctx context.Context, cancel context.CancelFunc) {
	for _, wire := range a.wires {
		go func(from queue.OutQueue, to queue.InQueue, task func(interface{}) (interface{}, error)) {
			defer func() {
				// 先にGraphを止めてからQueueを止める
				cancel()
				from.Close()
				to.Close()
			}()
		T:
			for {
				select {
				case <-ctx.Done():
					break T
				default:
					v, err := from.Pop(ctx)
					if err != nil {
						break T
					}
					r, err := task(v)
					if err != nil {
						break T
					}
					err = to.Push(ctx, r)
					if err != nil {
						break T
					}
				}
			}
		}(wire.from, wire.to, wire.task)
	}
}
