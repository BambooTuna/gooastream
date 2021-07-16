package std

import (
	"github.com/BambooTuna/gooastream/builder"
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
)

type (
	flowImpl struct {
		in        queue.Queue
		out       queue.Queue
		graphTree *stream.GraphTree
	}
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

var _ stream.Flow = (*flowImpl)(nil)
var _ builder.Balance = (*balanceImpl)(nil)
var _ builder.Merge = (*mergeImpl)(nil)

func NewBufferFlow(buffer int) stream.Flow {
	in := queue.NewQueueEmpty(buffer)
	out := queue.NewQueueEmpty(buffer)
	return &flowImpl{
		in:        in,
		out:       out,
		graphTree: stream.PassThrowGraph(in, out),
	}
}

/*
     +-------------------------------------------------------+
     | Resulting Flow                                       |
     |                                                      |
     |  +-------------+                  +---------------+  |
     |  |             |                  |               |  |
 I  ~~> | Sink        | [no-connection!] | Source        | ~~> O
     |  |             |                  |               |  |
     |  +-------------+                  +---------------+  |
     +------------------------------------------------------+

どちらかがクローズしても、その事はもう一方へ伝達されない
*/
func NewFlowFromSinkAndSource(sink stream.Sink, source stream.Source) stream.Flow {
	return &flowImpl{
		in:  sink.In(),
		out: source.Out(),
		graphTree: sink.GraphTree().
			Append(source.GraphTree()),
	}
}

func NewFlowFanout(flow stream.Flow, size int) []stream.Flow {
	flows := make([]stream.Flow, size)
	for i := 0; i < size; i++ {
		out := NewBufferFlow(0)
		flows[i] = &flowImpl{
			in:  flow.In(),
			out: out.Out(),
			graphTree: flow.GraphTree().
				Append(stream.PassThrowGraph(flow.Out(), out.In())).
				Append(out.GraphTree()),
		}
	}
	return flows
}

func (a *flowImpl) Via(flow stream.Flow) stream.Flow {
	return &flowImpl{
		in:  a.In(),
		out: flow.Out(),
		graphTree: a.GraphTree().
			Append(stream.PassThrowGraph(a.Out(), flow.In())).
			Append(flow.GraphTree()),
	}
}
func (a *flowImpl) To(sink stream.Sink) stream.Sink {
	return &sinkImpl{
		in: a.In(),
		graphTree: a.GraphTree().
			Append(stream.PassThrowGraph(a.Out(), sink.In())).
			Append(sink.GraphTree()),
	}
}
func (a *flowImpl) In() queue.Queue {
	return a.in
}
func (a *flowImpl) Out() queue.Queue {
	return a.out
}
func (a *flowImpl) GraphTree() *stream.GraphTree {
	return a.graphTree
}

func NewBalance(size int) builder.Balance {
	in := queue.NewQueueEmpty(0)
	outs := make([]queue.Queue, size)
	graphTree := stream.EmptyGraph()
	for i := 0; i < size; i++ {
		out := queue.NewQueueEmpty(0)
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

func NewMerge(size int) builder.Merge {
	ins := make([]queue.Queue, size)
	out := queue.NewQueueEmpty(0)
	graphTree := stream.EmptyGraph()
	for i := 0; i < size; i++ {
		in := queue.NewQueueEmpty(0)
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
