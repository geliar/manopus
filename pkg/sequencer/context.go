package sequencer

import (
	"context"
	"sync"
	"time"
)

type mergedContext struct {
	mu      sync.Mutex
	mainCtx context.Context
	ctx     context.Context
	done    chan struct{}
	err     error
}

func mergeContexts(mainCtx, ctx context.Context) context.Context {
	c := &mergedContext{mainCtx: mainCtx, ctx: ctx, done: make(chan struct{})}
	go c.run()
	return c
}

func (c *mergedContext) Done() <-chan struct{} {
	return c.done
}

func (c *mergedContext) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

func (c *mergedContext) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c *mergedContext) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

func (c *mergedContext) run() {
	var doneCtx context.Context
	select {
	case <-c.mainCtx.Done():
		doneCtx = c.mainCtx
	case <-c.ctx.Done():
		doneCtx = c.ctx
	case <-c.done:
		return
	}

	c.mu.Lock()
	if c.err != nil {
		c.mu.Unlock()
		return
	}
	c.err = doneCtx.Err()
	c.mu.Unlock()
	close(c.done)
}
