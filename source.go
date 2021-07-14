package gooastream

type (
	SourceRef interface {
		InQueue
		CloserQueue
	}
	Source interface {
		Via(Flow) Source
		To(Sink) Runnable

		CloseToSingleOutput
		Graph
	}
	sourceImpl struct {
		id     string
		out    Queue
		graphs graphs

		left interface{}
	}
)

func NewChannelSource(buffer int) (SourceRef, Source) {
	in := NewQueueEmpty(buffer)
	out := NewQueueEmpty(buffer)
	return in, &sourceImpl{
		id:  "channel source",
		out: out,
		graphs: graphs{
			&graph{
				id:   "channel source",
				from: in,
				to:   out,
				task: func(i interface{}) interface{} {
					return i
				},
			},
		},
	}
}

func NewSource(list []interface{}, buffer int) Source {
	in := NewQueueSlice(list)
	out := NewQueueEmpty(buffer)
	return &sourceImpl{
		id:  "source",
		out: out,
		graphs: graphs{
			&graph{
				id:   "source",
				from: in,
				to:   out,
				task: func(i interface{}) interface{} {
					return i
				},
			},
		},
	}
}
func (a *sourceImpl) Via(flow Flow) Source {
	a.graphs = append(a.graphs, &graph{
		id:   "source via",
		from: a.getOut(),
		to:   flow.getIn(),
		task: func(i interface{}) interface{} {
			return i
		},
	})
	a.graphs = append(a.graphs, flow.getGraphs()...)
	return &sourceImpl{
		out:    flow.getOut(),
		graphs: a.graphs,
	}
}
func (a *sourceImpl) To(sink Sink) Runnable {
	a.graphs = append(a.graphs, &graph{
		id:   "source to",
		from: a.getOut(),
		to:   sink.getIn(),
		task: func(i interface{}) interface{} {
			return i
		},
	})
	a.graphs = append(a.graphs, sink.getGraphs()...)
	return &runnableImpl{
		graphs: a.graphs,
		left:   a.getLeft(),
		right:  sink.getRight(),
	}
}
func (a *sourceImpl) getOut() Queue {
	return a.out
}
func (a *sourceImpl) getLeft() interface{} {
	return a.left
}
func (a *sourceImpl) getGraphs() graphs {
	return a.graphs
}
