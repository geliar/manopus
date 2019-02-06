package timer

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/geliar/manopus/pkg/input"

	"github.com/geliar/manopus/pkg/payload"
)

type Timer struct {
	created  int64
	id       int64
	name     string
	handlers []input.Handler
	stop     bool
	stopCh   chan struct{}
	mu       sync.RWMutex
}

func (c *Timer) Name() string {
	return c.name
}

func (c *Timer) Type() string {
	return connectorName
}

func (c *Timer) RegisterHandler(ctx context.Context, handler input.Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers = append(c.handlers, handler)
}

func (c *Timer) Send(ctx context.Context, response *payload.Response) map[string]interface{} {
	l := logger(ctx)
	l.Debug().
		Str("input_name", response.Request.Input).
		Str("input_event_id", response.Request.ID).
		Msg("Received Send event")
	if response.Data == nil {
		l.Error().Msg("Data field of request is empty")
		return nil
	}
	f, _ := response.Data["function"].(string)

	switch f {
	case "":
		l.Error().Msg("function field is empty")
		return nil
	case "timer":
		var d time.Duration
		switch v := response.Data["duration"].(type) {
		case int:
			d = time.Duration(v)
		case int64:
			d = time.Duration(v)
		case float64:
			d = time.Duration(v)
		case string:
			t, _ := strconv.Atoi(v)
			d = time.Duration(t)
		}
		if d <= 0 {
			l.Error().Msg("duration field is empty")
			return nil
		}
		id := c.getID()
		dur := time.Duration(d) * time.Second
		time.AfterFunc(dur, func() {
			c.sendEventToHandlers(ctx, &payload.Event{
				Input: serviceName,
				ID:    id,
				Data: map[string]interface{}{
					"timer_id": id,
					"now":      time.Now().Unix(),
				},
			})
		})
		return map[string]interface{}{
			"timer_id": id,
			"now":      time.Now().Unix(),
		}
	}
	return nil
}

func (c *Timer) Stop(ctx context.Context) {
	if !c.stop {
		c.stop = true
		close(c.stopCh)
	}
}

func (c *Timer) getID() string {
	id := atomic.AddInt64(&c.id, 1)
	return fmt.Sprintf("%s-%d-%d", c.name, c.created, id)
}

func (c *Timer) sendEventToHandlers(ctx context.Context, event *payload.Event) (response *payload.Response) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.handlers {
		resp := h(ctx, event)
		if resp != nil {
			response = new(payload.Response)
			response.Data = map[string]interface{}{
				"data": resp,
			}
			response.ID = event.ID
			response.Request = event
			response.Output = serviceName
		}
	}
	return
}

func (c *Timer) ticker(ctx context.Context, duration time.Duration) {
	l := logger(ctx)
	t := time.NewTicker(duration)
	for {
		select {
		case <-c.stopCh:
			l.Info().Msg("Ticker has been stopped")
			t.Stop()
			return
		case now := <-t.C:
			id := c.getID()
			c.sendEventToHandlers(ctx, &payload.Event{
				Input: serviceName,
				ID:    id,
				Data: map[string]interface{}{
					"ticker_id": id,
					"now":       now.Unix(),
				},
			})
		}
	}
}
