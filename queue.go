package gooastream

import (
	"context"
	"fmt"
)

type (
	Queue interface {
		OutQueue
		InQueue
		CloserQueue
	}
	OutQueue interface {
		Pop(context.Context) (interface{}, error)
	}
	InQueue interface {
		Push(context.Context, interface{})
	}
	CloserQueue interface {
		Close()
	}

	queueImpl struct {
		q Mailbox
	}
	queueSliceImpl struct {
		q Mailbox
	}
	queueSinkImpl struct {
		task func(interface{})
	}
)

var EmptyQueueError = fmt.Errorf("empty")

func NewQueueEmpty(size int) Queue {
	return &queueImpl{
		q: NewMailbox(size),
	}
}
func (a *queueImpl) Pop(ctx context.Context) (interface{}, error) {
	return a.q.DequeueOrWaitForElement(ctx)
}
func (a *queueImpl) Push(ctx context.Context, in interface{}) {
	a.q.EnqueueOrWaitForVacant(ctx, in)
}
func (a *queueImpl) Close() {
	a.q.Close()
}

func NewQueueSlice(list []interface{}) Queue {
	q := NewMailbox(len(list))
	go func() {
		for _, v := range list {
			q.EnqueueOrWaitForVacant(context.Background(), v)
		}
	}()
	return &queueSliceImpl{
		q: q,
	}
}

func (a *queueSliceImpl) Pop(ctx context.Context) (interface{}, error) {
	return a.q.DequeueOrWaitForElement(ctx)
}
func (a *queueSliceImpl) Push(ctx context.Context, in interface{}) {
	a.q.EnqueueOrWaitForVacant(ctx, in)
}
func (a *queueSliceImpl) Close() {
	a.q.Close()
}

func NewQueueSink(task func(interface{})) Queue {
	return &queueSinkImpl{
		task: task,
	}
}

func (a *queueSinkImpl) Pop(ctx context.Context) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, EmptyQueueError
	}
}
func (a *queueSinkImpl) Push(ctx context.Context, in interface{}) {
	select {
	case <-ctx.Done():
	default:
		a.task(in)
	}
}
func (a *queueSinkImpl) Close() {}
