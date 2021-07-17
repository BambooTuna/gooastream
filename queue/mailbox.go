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

	MailboxStrategy uint8

	mailboxImpl struct {
		mu       sync.RWMutex
		isClosed bool
		queue    chan interface{}
		strategy MailboxStrategy
	}

	mailboxInfiniteImpl struct {
		mu       sync.RWMutex
		isClosed bool
		element  interface{}
	}
)

func NewMailbox() Mailbox {
	return NewMailboxWithBuffer(defaultQueueOption())
}

func NewMailboxWithBuffer(option queueOption) Mailbox {
	return NewMailboxWithStrategy(option.size, option.strategy)
}

func NewMailboxWithStrategy(size int, strategy MailboxStrategy) Mailbox {
	return &mailboxImpl{
		queue:    make(chan interface{}, size),
		strategy: strategy,
	}
}

func (a *mailboxImpl) EnqueueOrWaitForVacant(ctx context.Context, in interface{}) error {
	a.mu.RLock()
	if a.isClosed {
		a.mu.RUnlock()
		return fmt.Errorf("queue is closed")
	}
	select {
	case <-ctx.Done():
		a.mu.RUnlock()
		return ctx.Err()
	case a.queue <- in:
		a.mu.RUnlock()
		return nil
	default:
		switch a.strategy {
		case Backpressure:
			a.mu.RUnlock()
			a.queue <- in
			return nil
		case DropNew:
			a.mu.RUnlock()
			return nil
		case DropHead:
			_ = <-a.queue
			a.queue <- in
			a.mu.RUnlock()
			return nil
		case DropBuffer:
		T:
			for {
				select {
				case _ = <-a.queue:
				default:
					break T
				}
			}
			a.mu.RUnlock()
		case Fail:
			return fmt.Errorf("buffer is full")
		default:
			return fmt.Errorf("unknown strategy")
		}
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
	case out, ok := <-a.queue:
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
