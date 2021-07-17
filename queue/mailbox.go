package queue

import (
	"context"
	"fmt"
	"sync"
)

type (
	Mailbox interface {
		EnqueueOrWaitForVacant(context.Context, interface{}) error
		DequeueOrWaitForElement(context.Context) (interface{}, error)
		Close()
	}

	mailboxImpl struct {
		mu       sync.RWMutex
		isClosed bool
		queue    chan interface{}
	}

	mailboxInfiniteImpl struct {
		mu       sync.RWMutex
		isClosed bool
		element  interface{}
	}
)

func NewMailbox(size int) Mailbox {
	return &mailboxImpl{
		queue: make(chan interface{}, size),
	}
}

func (a *mailboxImpl) EnqueueOrWaitForVacant(ctx context.Context, in interface{}) error {
	a.mu.RLock()
	if a.isClosed {
		return fmt.Errorf("queue is closed")
	}
	a.mu.RUnlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		a.queue <- in
		return nil
	}
}
func (a *mailboxImpl) DequeueOrWaitForElement(ctx context.Context) (interface{}, error) {
	a.mu.RLock()
	if a.isClosed {
		return nil, fmt.Errorf("queue is closed")
	}
	a.mu.RUnlock()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		out, ok := <-a.queue
		if !ok {
			return nil, fmt.Errorf("closed")
		}
		return out, nil
	}
}
func (a *mailboxImpl) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.isClosed {
		a.isClosed = true
		close(a.queue)
		// Garbage disposal
		go func() {
			for {
				_, ok := <-a.queue
				if !ok {
					break
				}
			}
		}()
	}
}

func NewMailboxInfinite(element interface{}) Mailbox {
	return &mailboxInfiniteImpl{
		element: element,
	}
}

func (a *mailboxInfiniteImpl) EnqueueOrWaitForVacant(ctx context.Context, in interface{}) error {
	a.mu.RLock()
	if a.isClosed {
		return fmt.Errorf("queue is closed")
	}
	a.mu.RUnlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		a.element = in
		return nil
	}
}
func (a *mailboxInfiniteImpl) DequeueOrWaitForElement(ctx context.Context) (interface{}, error) {
	a.mu.RLock()
	if a.isClosed {
		return nil, fmt.Errorf("queue is closed")
	}
	a.mu.RUnlock()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return a.element, nil
	}
}
func (a *mailboxInfiniteImpl) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.isClosed {
		a.isClosed = true
	}
}
