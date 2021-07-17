package queue

import (
	"context"
	"fmt"
	"sync"
)

type (
	Queue interface {
		OutQueue
		InQueue
	}
	OutQueue interface {
		Pop(context.Context) (interface{}, error)
		CloserQueue
	}
	InQueue interface {
		Push(context.Context, interface{}) error
		CloserQueue
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
		mu       sync.RWMutex
		isClosed bool
		task     func(interface{}) error
	}
)

var EmptyQueueError = fmt.Errorf("empty")
var ClosedQueueError = fmt.Errorf("sink is closed")

func NewQueueEmpty(options ...Option) Queue {
	option := defaultQueueOption()
	for _, v := range options {
		v(&option)
	}
	return &queueImpl{
		q: NewMailboxWithBuffer(option),
	}
}
func NewInfiniteElement(element interface{}) Queue {
	return &queueImpl{
		q: NewMailboxInfinite(element),
	}
}
func (a *queueImpl) Pop(ctx context.Context) (interface{}, error) {
	return a.q.DequeueOrWaitForElement(ctx)
}
func (a *queueImpl) Push(ctx context.Context, in interface{}) error {
	return a.q.EnqueueOrWaitForVacant(ctx, in)
}
func (a *queueImpl) Close() {
	a.q.Close()
}

func NewQueueSlice(list []interface{}, options ...Option) Queue {
	option := defaultQueueOption()
	for _, v := range options {
		v(&option)
	}
	q := NewMailboxWithBuffer(option)
	go func() {
		for _, v := range list {
			_ = q.EnqueueOrWaitForVacant(context.Background(), v)
		}
	}()
	return &queueSliceImpl{
		q: q,
	}
}

func (a *queueSliceImpl) Pop(ctx context.Context) (interface{}, error) {
	return a.q.DequeueOrWaitForElement(ctx)
}
func (a *queueSliceImpl) Push(ctx context.Context, in interface{}) error {
	return a.q.EnqueueOrWaitForVacant(ctx, in)
}
func (a *queueSliceImpl) Close() {
	a.q.Close()
}

func NewQueueSink(task func(interface{}) error) Queue {
	return &queueSinkImpl{
		task: task,
	}
}

func (a *queueSinkImpl) Pop(ctx context.Context) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ClosedQueueError
	default:
		return nil, EmptyQueueError
	}
}
func (a *queueSinkImpl) Push(ctx context.Context, in interface{}) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.isClosed {
		return ClosedQueueError
	}
	select {
	case <-ctx.Done():
		return ClosedQueueError
	default:
		return a.task(in)
	}
}
func (a *queueSinkImpl) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.isClosed = true
}
