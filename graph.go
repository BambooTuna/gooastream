package gooastream

import (
	"context"
)

type (
	graph struct {
		wires []*wire
	}

	Graph interface {
		getGraph() *graph
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
		from OutQueue
		to   InQueue
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

func emptyGraph() *graph {
	return &graph{
		wires: []*wire{},
	}
}

func passThrowGraph(from OutQueue, to InQueue) *graph {
	return &graph{
		wires: []*wire{
			{
				from: from,
				to:   to,
				task: emptyTaskFunc,
			},
		},
	}
}

func mapGraph(from OutQueue, to InQueue, f func(interface{}) (interface{}, error)) *graph {
	return &graph{
		wires: []*wire{
			{
				from: from,
				to:   to,
				task: f,
			},
		},
	}
}

func (a *graph) Append(child *graph) *graph {
	return &graph{wires: append(a.wires, child.wires...)}
}

// non blocking
func (a *graph) Run(ctx context.Context, cancel context.CancelFunc) {
	for _, wire := range a.wires {
		go func(from OutQueue, to InQueue, task func(interface{}) (interface{}, error)) {
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
