package gooastream

import "context"

type (
	Graph interface {
		getGraphs() graphs
	}
	graph struct {
		id   string
		from OutQueue
		to   InQueue
		task func(interface{}) (interface{}, error)
	}
	graphs []*graph
)

var emptyTaskFunc = func(i interface{}) (interface{}, error) {
	return i, nil
}

func (a graphs) run(ctx context.Context, cancel func()) {
	for _, v := range a {
		go v.run(ctx, cancel)
	}
}
func (a *graph) run(ctx context.Context, cancel func()) {
	defer func() {
		// 先にGraphを止めてからQueueを止める
		cancel()
		a.from.Close()
		a.to.Close()
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
			r, err := a.task(v)
			if err != nil {
				break T
			}
			err = a.to.Push(ctx, r)
			if err != nil {
				break T
			}
		}
	}
}
