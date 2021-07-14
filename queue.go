package gooastream

import (
	"context"
	"fmt"
	"sync"
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
		Push(context.Context, interface{}) error
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

func NewQueueEmpty(size int) Queue {
	return &queueImpl{
		q: NewMailbox(size),
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

func NewQueueSlice(list []interface{}) Queue {
	q := NewMailbox(len(list))
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
