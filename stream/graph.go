package stream

import (
	"context"
	"errors"
	"fmt"
	"github.com/BambooTuna/gooastream/queue"
	"time"
)

// Mat
// unused
type Mat uint8

const (
	MatNone Mat = iota
	MatLeft
	MatRight
	MatBoth
)

// GraphTree
// A structure that represents the connection between queue.OutQueue
type GraphTree struct {
	wires []*wire
}

type wire struct {
	from     queue.OutQueue
	to       queue.InQueue
	throttle time.Duration
	task     func(interface{}) (interface{}, error)
}

var PassPermissionError = fmt.Errorf("pass permission error")
var emptyTaskFunc = func(i interface{}) (interface{}, error) {
	return i, nil
}

/*
	EmptyGraph
	Create a empty GraphTree.
*/
func EmptyGraph() *GraphTree {
	return &GraphTree{
		wires: []*wire{},
	}
}

/*
	PassThrowGraph
	Create a GraphTree with from queue.OutQueue and to queue.InQueue.
	Just pass the data between from queue.OutQueue and to queue.InQueue.
*/
func PassThrowGraph(from queue.OutQueue, to queue.InQueue) *GraphTree {
	return ThrottleGraph(from, to, 0)
}

/*
	ThrottleGraph
	Create a GraphTree with from queue.OutQueue, to queue.InQueue and throttle.
	Once in a certain period, pass the data between from queue.OutQueue and to queue.InQueue.
*/
func ThrottleGraph(from queue.OutQueue, to queue.InQueue, throttle time.Duration) *GraphTree {
	return &GraphTree{
		wires: []*wire{
			{
				from:     from,
				to:       to,
				throttle: throttle,
				task:     emptyTaskFunc,
			},
		},
	}
}

/*
	MapGraph
	Create a GraphTree with from queue.OutQueue, to queue.InQueue and map function.
	Pass the result of passing the data of from queue.OutQueue through the function to to queue.InQueue.
*/
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

/*
	Append
	Mix self and another GraphTree to make a new GraphTree.
*/
func (a *GraphTree) Append(child *GraphTree) *GraphTree {
	return &GraphTree{wires: append(a.wires, child.wires...)}
}

/*
	Add
	Mix child GraphTree with self
*/
func (a *GraphTree) Add(child *GraphTree) {
	a.wires = append(a.wires, child.wires...)
}

/*
	Run
	Run GraphTree with context.Context and context.CancelFunc.
	context.CancelFunc is called if some error occurs inside
	non blocking.
*/
func (a *GraphTree) Run(ctx context.Context, cancel context.CancelFunc) {
	for _, wire := range a.wires {
		go func(from queue.OutQueue, to queue.InQueue, throttle time.Duration, task func(interface{}) (interface{}, error)) {
			defer func() {
				// 先にGraphを止めてからQueueを止める
				cancel()
				from.Close()
				to.Close()
			}()
			if throttle > 0 {
				throttler(ctx, from, to, throttle, task)
			} else {
				passThrow(ctx, from, to, task)
			}
		}(wire.from, wire.to, wire.throttle, wire.task)
	}
}

func passThrow(ctx context.Context, from queue.OutQueue, to queue.InQueue, task func(interface{}) (interface{}, error)) {
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
				if errors.Is(err, PassPermissionError) {
					continue
				}
				break T
			}
			err = to.Push(ctx, r)
			if err != nil {
				break T
			}
		}
	}
}

func throttler(ctx context.Context, from queue.OutQueue, to queue.InQueue, throttle time.Duration, task func(interface{}) (interface{}, error)) {
	ticker := time.NewTicker(throttle)
	defer ticker.Stop()
T:
	for range ticker.C {
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
				if errors.Is(err, PassPermissionError) {
					continue
				}
				break T
			}
			err = to.Push(ctx, r)
			if err != nil {
				break T
			}
		}
	}
}
