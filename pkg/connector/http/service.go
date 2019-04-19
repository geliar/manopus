package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
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

// Name returns name of connector
func (c *HTTP) Name() string {
	return c.name
}

// Type returns type of connector
func (c *HTTP) Type() string {
	return connectorName
}

// RegisterHandler with connector
func (c *HTTP) RegisterHandler(ctx context.Context, handler input.Handler) {
	c.Lock()
	defer c.Unlock()
	c.handlers = append(c.handlers, handler)
}

// Send response with connector
func (c *HTTP) Send(ctx context.Context, response *payload.Response) map[string]interface{} {
	l := logger(ctx)
	l.Debug().
		Str("input_name", response.Request.Input).
		Str("input_event_id", response.Request.ID).
		Msg("Received Send event")
	return nil
}

// Stop connector
func (c *HTTP) Stop(ctx context.Context) {
	c.stop = true
}

func (c *HTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := logger(r.Context())
	_ = r.ParseForm()
	var buf []byte
	func() {
		var err error
		buf, err = ioutil.ReadAll(r.Body)
		defer func() { _ = r.Body.Close() }()
		if err != nil {
			l.Error().Err(err).Msg("Cannot read HTTP request body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}()
	e := payload.Event{
		ID:    c.getID(),
		Type:  RequestTypeHTTPRequest,
		Input: c.name,
		Data: RequestHTTPRequest{
			Method:      r.Method,
			Host:        r.Host,
			RemoteAddr:  r.RemoteAddr,
			Uri:         r.RequestURI,
			Path:        r.URL.Path,
			Form:        r.Form,
			ContentType: r.Header.Get("Content-Type"),
			Referer:     r.Referer(),
			UserAgent:   r.UserAgent(),
			Headers:     r.Header,
			Body:        string(buf),
		},
	}
	switch r.Header.Get("Content-Type") {
	case "application/json":
		var v map[string]interface{}
		err := json.Unmarshal(buf, &v)
		if err != nil {
			l.Error().Err(err).Msg("Cannot unmarshal JSON request")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		e.Type = RequestTypeHTTPJSONRequest
		e.Data = RequestHTTPJSONRequest{
			RequestHTTPRequest: e.Data.(RequestHTTPRequest),
			JSON:               v,
		}
	}
	response := c.sendEventToHandlers(r.Context(), &e)
	if response == nil {
		return
	}
	if response.Data["data"] == nil {
		return
	}
	switch v := response.Data["data"].(type) {
	case string:
		buf = []byte(v)
	case []byte:
		buf = v
	case map[string]interface{}, []string, []int, []interface{}:
		var err error
		buf, err = json.Marshal(response.Data["data"])
		if err != nil {
			l.Error().
				Err(err).
				Msg("Error when marshaling response to JSON")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
	default:
		l.Error().Msgf("Unknown type of 'data' field of response %T", response.Data["data"])
		return
	}

	if contentType, ok := response.Data["httpContentType"].(string); ok {
		w.Header().Set("Content-Type", contentType)
	}

	switch v := response.Data["http_code"].(type) {
	case string:
		if code, err := strconv.Atoi(v); err != nil {
			w.WriteHeader(code)
		}
	case int:
		w.WriteHeader(v)
	case int64:
		w.WriteHeader(int(v))
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

func (c *HTTP) getID() string {
	id := atomic.AddInt64(&c.id, 1)
	return fmt.Sprintf("%s-%d-%d", c.name, c.created, id)
}
