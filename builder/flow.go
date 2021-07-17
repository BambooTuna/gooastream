package builder

import (
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
)

type (
	balanceImpl struct {
		in        queue.Queue
		out       []queue.Queue
		graphTree *stream.GraphTree
	}
	mergeImpl struct {
		in        []queue.Queue
		out       queue.Queue
		graphTree *stream.GraphTree
	}
)

var _ Balance = (*balanceImpl)(nil)
var _ Merge = (*mergeImpl)(nil)

func NewBroadcast(size int, options ...queue.Option) Balance {
	in := queue.NewQueueEmpty(options...)
	outs := make([]queue.Queue, size)
	broadcasts := make([]queue.InQueue, size)
	for i := 0; i < size; i++ {
		out := queue.NewQueueEmpty(options...)
		outs[i] = out
		broadcasts[i] = out
	}
	return &balanceImpl{
		in:        in,
		out:       outs,
		graphTree: stream.BroadcastGraph(in, broadcasts),
	}
}

func NewBalance(size int, options ...queue.Option) Balance {
	in := queue.NewQueueEmpty(options...)
	outs := make([]queue.Queue, size)
	graphTree := stream.EmptyGraph()
	for i := 0; i < size; i++ {
		out := queue.NewQueueEmpty(options...)
		outs[i] = out
		graphTree.Add(stream.PassThrowGraph(in, out))
	}
	return &balanceImpl{
		in:        in,
		out:       outs,
		graphTree: graphTree,
	}
}

func (a *balanceImpl) In() queue.Queue {
	return a.in
}
func (a *balanceImpl) Out() []queue.Queue {
	return a.out
}
func (a *balanceImpl) GraphTree() *stream.GraphTree {
	return a.graphTree
}

func NewMerge(size int, options ...queue.Option) Merge {
	ins := make([]queue.Queue, size)
	out := queue.NewQueueEmpty(options...)
	graphTree := stream.EmptyGraph()
	for i := 0; i < size; i++ {
		in := queue.NewQueueEmpty(options...)
		ins[i] = in
		graphTree.Add(stream.PassThrowGraph(in, out))
	}
	return &mergeImpl{
		in:        ins,
		out:       out,
		graphTree: graphTree,
	}
}

func (a *mergeImpl) In() []queue.Queue {
	return a.in
}
func (a *mergeImpl) Out() queue.Queue {
	return a.out
}
func (a *mergeImpl) GraphTree() *stream.GraphTree {
	return a.graphTree
}
