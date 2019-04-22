package bitbucket

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"

	cbitbucket "github.com/ktrysmt/go-bitbucket"
	"github.com/rs/zerolog/hlog"
	whbitbucket "gopkg.in/go-playground/webhooks.v5/bitbucket"

	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/payload"
)

//Bitbucket implementation of the Bitbucket connector
type Bitbucket struct {
	created  int64
	id       int64
	name     string
	handlers []input.Handler
	stop     bool
	stopCh   chan struct{}
	mu       sync.RWMutex
	hook     *whbitbucket.Webhook
	client   *cbitbucket.Client
}

// Name returns name of the connector
func (c *Bitbucket) Name() string {
	return c.name
}

// Type returns type of connector
func (c *Bitbucket) Type() string {
	return serviceName
}

// RegisterHandler registers event handler with connector
func (c *Bitbucket) RegisterHandler(ctx context.Context, handler input.Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers = append(c.handlers, handler)
}

// Send sends response with connector
func (c *Bitbucket) Send(ctx context.Context, response *payload.Response) map[string]interface{} {
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
	case "pull_request_merge":
		po := new(cbitbucket.PullRequestsOptions)
		po.Owner = response.Data["owner"].(string)
		po.RepoSlug = response.Data["owner"].(string)
		po.ID = strconv.Itoa(int(response.Data["id"].(int64)))
		po.Message = response.Data["message"].(string)
		_, err := c.client.Repositories.PullRequests.Merge(po)
		if err != nil {
			l.Error().Err(err).Msg("Error on pull_request_merge")
		}
		return nil
	}
	return nil
}

func (c *Bitbucket) webhookHandler(w http.ResponseWriter, r *http.Request) {
	tl := hlog.FromRequest(r)
	ctx := tl.WithContext(context.Background())
	_ = ctx
	l := logger(ctx)
	data, err := c.hook.Parse(r, whbitbucket.PullRequestCreatedEvent)
	if err != nil {
		l.Error().Err(err).Msg("Error parsing event")
		if err == whbitbucket.ErrEventNotFound {

		}
	}
	//spew.Dump(data)
	event := new(payload.Event)
	event.ID = c.getID()
	event.Input = c.name
	switch v := data.(type) {
	case whbitbucket.PullRequestCreatedPayload:
		event.Type = requestTypePullRequestCreated
		event.Data = requestPullRequestCreated{
			PullRequestCreatedPayload: v,
		}
		l.Debug().Msg("Pull request created event")
	case whbitbucket.PullRequestApprovedPayload:
		event.Type = requestTypePullRequestApproved
		event.Data = requestPullRequestApproved{
			PullRequestApprovedPayload: v,
		}

		l.Debug().Msg("Pull request approved event")
	case whbitbucket.RepoPushPayload:
		event.Type = requestTypeRepoPush
		event.Data = requestPush{
			RepoPushPayload: v,
		}

		l.Debug().Msg("Repo push event")
	}
	if event.Data != nil {
		c.sendEventToHandlers(ctx, event)
	}
}

// Stop connector
func (c *Bitbucket) Stop(ctx context.Context) {
	if !c.stop {
		c.stop = true
		close(c.stopCh)
	}
}

func (c *Bitbucket) getID() string {
	id := atomic.AddInt64(&c.id, 1)
	return fmt.Sprintf("%s-%d-%d", c.name, c.created, id)
}

func (c *Bitbucket) sendEventToHandlers(ctx context.Context, event *payload.Event) (response *payload.Response) {
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
