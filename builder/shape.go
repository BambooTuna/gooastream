package builder

import (
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
)

type (
	Inlet struct {
		connected bool
		in        queue.Queue
	}
	Outlet struct {
		connected bool
		graphTree *stream.GraphTree
		out       queue.Queue
	}

	Shape       interface{}
	SourceShape interface {
		Shape
		Out() *Outlet
	}
	FlowShape interface {
		Shape
		In() *Inlet
		Out() *Outlet
	}
	SinkShape interface {
		Shape
		In() *Inlet
	}
	FanOutShape interface {
		Shape
		In() *Inlet
		Out() []*Outlet
	}
	FanInShape interface {
		Shape
		In() []*Inlet
		Out() *Outlet
	}
)

// GraphBuilder is useful for creating complex processing graphs.
type GraphBuilder struct {
	graphTree *stream.GraphTree
}

type (
	sourceShape struct {
		outlet *Outlet
	}
	flowShape struct {
		inlet  *Inlet
		outlet *Outlet
	}
	fanOutShape struct {
		inlet  *Inlet
		outlet []*Outlet
	}
	fanInShape struct {
		inlet  []*Inlet
		outlet *Outlet
	}
	sinkShape struct {
		inlet *Inlet
	}
)

//var _ Shape = (*FlowShape)(nil)

func NewGraphBuilder() *GraphBuilder {
	return &GraphBuilder{
		graphTree: stream.EmptyGraph(),
	}
}
func (a *GraphBuilder) AddSource(stream stream.Source) SourceShape {
	a.graphTree.Add(stream.GraphTree())
	return newSourceShape(a.graphTree, stream)
}
func (a *GraphBuilder) AddFlow(stream stream.Flow) FlowShape {
	a.graphTree.Add(stream.GraphTree())
	return newFlowShape(a.graphTree, stream)
}
func (a *GraphBuilder) AddBalance(stream Balance) FanOutShape {
	a.graphTree.Add(stream.GraphTree())
	return newFanOutShape(a.graphTree, stream)
}
func (a *GraphBuilder) AddMerge(stream Merge) FanInShape {
	a.graphTree.Add(stream.GraphTree())
	return newFanInShape(a.graphTree, stream)
}
func (a *GraphBuilder) AddSink(stream stream.Sink) SinkShape {
	a.graphTree.Add(stream.GraphTree())
	return newSinkShape(stream)
}
func (a *GraphBuilder) ToRunnable() stream.Runnable {
	// TODO check all added stream is connected
	return stream.NewRunnable(a.graphTree)
}
func (a *GraphBuilder) ToSource(out *Outlet) stream.Source {
	// TODO check all added stream is connected
	return stream.BuildSource(out.out, a.graphTree)
}
func (a *GraphBuilder) ToFlow(in *Inlet, out *Outlet) stream.Flow {
	// TODO check all added stream is connected
	return stream.BuildFlow(in.in, out.out, a.graphTree)
}
func (a *GraphBuilder) ToSink(in *Inlet) stream.Sink {
	// TODO check all added stream is connected
	return stream.BuildSink(in.in, a.graphTree)
}

func (a *Outlet) Wire(inlet *Inlet) {
	if a.connected || inlet.connected {
		panic("outlet or inlet is already connected")
	}
	a.connected = true
	inlet.connected = true
	a.graphTree.Add(stream.PassThrowGraph(a.out, inlet.in))
}

func newSourceShape(graphTree *stream.GraphTree, source stream.Source) SourceShape {
	return &sourceShape{
		outlet: &Outlet{
			graphTree: graphTree,
			out:       source.Out(),
		},
	}
}
func (a *sourceShape) Out() *Outlet {
	return a.outlet
}

func newFlowShape(graphTree *stream.GraphTree, flow stream.Flow) FlowShape {
	return &flowShape{
		inlet: &Inlet{in: flow.In()},
		outlet: &Outlet{
			graphTree: graphTree,
			out:       flow.Out(),
		},
	}
}
func (a *flowShape) In() *Inlet {
	return a.inlet
}
func (a *flowShape) Out() *Outlet {
	return a.outlet
}

func newFanOutShape(graphTree *stream.GraphTree, flow Balance) FanOutShape {
	outlets := make([]*Outlet, len(flow.Out()))
	for i, v := range flow.Out() {
		outlets[i] = &Outlet{
			graphTree: graphTree,
			out:       v,
		}
	}
	return &fanOutShape{
		inlet:  &Inlet{in: flow.In()},
		outlet: outlets,
	}
}
func (a *fanOutShape) In() *Inlet {
	return a.inlet
}
func (a *fanOutShape) Out() []*Outlet {
	return a.outlet
}

func newFanInShape(graphTree *stream.GraphTree, flow Merge) FanInShape {
	inlets := make([]*Inlet, len(flow.In()))
	for i, v := range flow.In() {
		inlets[i] = &Inlet{
			in: v,
		}
	}
	return &fanInShape{
		inlet: inlets,
		outlet: &Outlet{
			graphTree: graphTree,
			out:       flow.Out(),
		},
	}
}
func (a *fanInShape) In() []*Inlet {
	return a.inlet
}
func (a *fanInShape) Out() *Outlet {
	return a.outlet
}

func newSinkShape(sink stream.Sink) SinkShape {
	return &sinkShape{
		inlet: &Inlet{
			in: sink.In(),
		},
	}
}
func (a *sinkShape) In() *Inlet {
	return a.inlet
}
