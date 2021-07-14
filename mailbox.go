package gooastream

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

	mailboxImp struct {
		mu       sync.Mutex
		isClosed bool
		queue    chan interface{}
	}
)

func NewMailbox(size int) Mailbox {
	return &mailboxImp{
		queue: make(chan interface{}, size),
	}
}

func (a *mailboxImp) EnqueueOrWaitForVacant(ctx context.Context, in interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		a.queue <- in
		return nil
	}
}
func (a *mailboxImp) DequeueOrWaitForElement(ctx context.Context) (interface{}, error) {
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
func (a *mailboxImp) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.isClosed {
		a.isClosed = true
		close(a.queue)
	}
}
