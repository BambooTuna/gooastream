package builder

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
)

type (
	GraphTree interface {
		Append(GraphTree) GraphTree
		Run(ctx context.Context, cancel context.CancelFunc)
		getWires() []*wire
	}

	GraphNode interface {
		GraphTree() GraphTree
	}

	Mat uint8
	//
	//TraversalBuilder interface {
	//	Append(TraversalBuilder, Mat) TraversalBuilder
	//	Run(ctx context.Context, cancel context.CancelFunc)
	//
	//	Inlet() Inlet
	//	Outlet() Outlet
	//	Wires() []*wire
	//	Inlets() int
	//	Outlets() int
	//}

	wire struct {
		from queue.OutQueue
		to   queue.InQueue
		task func(interface{}) (interface{}, error)
	}

	graphTree struct {
		wires []*wire
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

func EmptyGraph() GraphTree {
	return &graphTree{
		wires: []*wire{},
	}
}

func PassThrowGraph(from queue.OutQueue, to queue.InQueue) GraphTree {
	return &graphTree{
		wires: []*wire{
			{
				from: from,
				to:   to,
				task: emptyTaskFunc,
			},
		},
	}
}

func MapGraph(from queue.OutQueue, to queue.InQueue, f func(interface{}) (interface{}, error)) GraphTree {
	return &graphTree{
		wires: []*wire{
			{
				from: from,
				to:   to,
				task: f,
			},
		},
	}
}

func (a *graphTree) Append(child GraphTree) GraphTree {
	return &graphTree{wires: append(a.wires, child.getWires()...)}
}

// non blocking
func (a *graphTree) Run(ctx context.Context, cancel context.CancelFunc) {
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

func (a *graphTree) getWires() []*wire {
	return a.wires
}
