package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/geliar/manopus/pkg/payload"

	"github.com/geliar/manopus/pkg/input"
)

type HTTP struct {
	created  int64
	id       int64
	name     string
	handlers []input.Handler
	stop     bool
	sync.RWMutex
}

func (*HTTP) validate() error {
	return nil
}

func (c *HTTP) Name() string {
	return c.name
}

func (c *HTTP) Type() string {
	return connectorName
}

func (c *HTTP) RegisterHandler(ctx context.Context, handler input.Handler) {
	c.Lock()
	defer c.Unlock()
	c.handlers = append(c.handlers, handler)
}

func (c *HTTP) Send(ctx context.Context, response *payload.Response) {
	l := logger(ctx)
	l.Debug().
		Str("input_name", response.Request.Input).
		Str("input_event_id", response.Request.ID).
		Msg("Received Send event")
}

func (c *HTTP) Stop(ctx context.Context) {
	c.stop = true
}

func (c *HTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := logger(r.Context())
	_ = r.ParseForm()
	e := payload.Event{
		ID:    c.getID(),
		Type:  connectorName,
		Input: c.name,
		Data: map[string]interface{}{
			"http_method":       r.Method,
			"http_uri":          r.RequestURI,
			"http_form":         r.Form,
			"http_content_type": r.Header.Get("Content-Type"),
		},
	}
	c.RLock()
	defer c.RUnlock()
	response := c.sendEventToHandlers(r.Context(), &e)
	if response == nil {
		return
	}
	var buf []byte
	switch response.Encoding {
	case "none", "plain":
		switch v := response.Data["data"].(type) {
		case string:
			buf = []byte(v)
		case []byte:
			buf = v
		default:
			l.Error().Msg("Unknown type of 'data' field of response")
			return
		}
	default:
		var err error
		buf, err = json.Marshal(response.Data)
		if err != nil {
			l.Error().
				Err(err).
				Msg("Error when marshaling response to JSON")
		}
	}
	_, err := w.Write(buf)
	if err != nil {
		l.Error().
			Err(err).
			Msg("Cannot send HTTP response")
	}
}

func (c *HTTP) sendEventToHandlers(ctx context.Context, event *payload.Event) (response *payload.Response) {
	c.RLock()
	defer c.RUnlock()
	for _, h := range c.handlers {
		resp := h(ctx, event)
		if resp != nil {
			response = resp
		}
	}
	return
}

func (c *HTTP) getID() string {
	id := atomic.AddInt64(&c.id, 1)
	return fmt.Sprintf("%s-%d-%d", c.name, c.created, id)
}
