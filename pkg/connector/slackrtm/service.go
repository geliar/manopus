package slackrtm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/output"

	"github.com/nlopes/slack"
)

type SlackRTM struct {
	created  int64
	id       int64
	name     string
	debug    bool
	token    string
	channels []string
	online   struct {
		Channels []slack.Channel
		Users    []slack.User
		User     slack.UserDetails
	}
	botIcon  slack.Icon
	rtm      *slack.RTM
	handlers []input.Handler
	sync.RWMutex
}

func (*SlackRTM) validate() error {
	return nil
}

func (c *SlackRTM) Name() string {
	return c.name
}

func (c *SlackRTM) Type() string {
	return connectorName
}

func (c *SlackRTM) RegisterHandler(ctx context.Context, handler input.Handler) {
	l := logger(ctx)
	l.Info().Msg("Registered new handler")
	c.Lock()
	defer c.Unlock()
	c.handlers = append(c.handlers, handler)
}

func (c *SlackRTM) Send(ctx context.Context, response *output.Response) {
	l := logger(ctx)
	if response.Type == "callback" {
		if response.Request.Input != c.Name() {
			l.Error().Msg("Cannot process callback response from different input")
			return
		}
		ch, ok := response.Request.Data["channel_id"]
		if !ok {
			l.Error().Msg("Cannot find `channel_id` field in request")
			return
		}
		chid, ok := ch.(string)
		if !ok {
			l.Error().Msg("Field `channel_id` should be string")
			return
		}
		c.sendToChannel(ctx, chid, response.Data.(string))
	}
}

func (c *SlackRTM) Stop(ctx context.Context) {
	if c.rtm != nil {
		_ = c.rtm.Disconnect()
	}
}

func (c *SlackRTM) serve(ctx context.Context) {
	l := logger(ctx)
	go c.rtm.ManageConnection()
	for msg := range c.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			l.Debug().Msgf("Infos: %v", ev.Info)
			l.Debug().Msgf("Connection counter: %d", ev.ConnectionCount)
			l.Debug().Msgf("Current user: %s (ID: %s)", ev.Info.User.Name, ev.Info.User.ID)
			l.Debug().Msgf("Current team: %s (ID: %s)", ev.Info.Team.Name, ev.Info.Team.ID)
			c.Lock()
			c.online.User = *ev.Info.User
			c.Unlock()

			c.updateChannels(ctx)
			c.updateUsers(ctx)

		case *slack.MessageEvent:
			l.Debug().Msgf("Message: %v\n", ev)
			// Only text messages from real users
			if ev.User != "" && ev.SubType == "" {
				e := &input.Event{
					Input: c.name,
					Type:  connectorName,
					ID:    c.getID(),
					Data: map[string]interface{}{
						"user_id":       ev.User,
						"user_name":     c.getUserByID(ctx, ev.User).Name,
						"user_realname": c.getUserByID(ctx, ev.User).RealName,
						"channel_id":    ev.Channel,
						"channel_name":  c.getChannelByID(ctx, ev.Channel).Name,
						"message":       ev.Text,
						"mentioned":     strings.Contains(ev.Text, fmt.Sprintf("<@%s>", c.online.User.ID)),
						"direct":        strings.HasPrefix(ev.Channel, "D"),
					},
				}
				c.sendEventToHandlers(ctx, e)
			}

		case *slack.RTMError:
			l.Debug().Err(ev).Msgf("Error: %s\n", ev.Error())
		case *slack.InvalidAuthEvent:
			l.Debug().Msgf("Invalid credentials")
			_ = c.rtm.Disconnect()
			return
		case *slack.DisconnectedEvent:
			l.Debug().Msgf("Disconnected event received. Stopping connector.")
			return
		default:
		}
	}
}

func (c *SlackRTM) sendEventToHandlers(ctx context.Context, event *input.Event) {
	c.RLock()
	defer c.RUnlock()
	for _, h := range c.handlers {
		go h(ctx, event)
	}
}

func (c *SlackRTM) updateChannels(ctx context.Context) {
	l := logger(ctx)
	l.Debug().Msg("Updating Slack channels")
	var resChannels []slack.Channel
	var cursor string
	for {
		channels, cur, err := c.rtm.GetConversationsContext(ctx,
			&slack.GetConversationsParameters{
				Cursor:          cursor,
				Limit:           1000,
				ExcludeArchived: "true",
				Types:           []string{"public_channel", "private_channel"}})
		if err != nil {
			l.Error().Err(err).Msg("Error when updating channels")
			return
		}
		resChannels = append(resChannels, channels...)
		if cur == "" {
			break
		}
		cursor = cur
	}
	c.Lock()
	c.online.Channels = resChannels
	c.Unlock()
	l.Debug().Msgf("Found %d channels", len(resChannels))
}

func (c *SlackRTM) updateUsers(ctx context.Context) {
	l := logger(ctx)
	l.Debug().Msg("Updating Slack users")
	users, err := c.rtm.GetUsersContext(ctx)
	if err != nil {
		l.Error().Err(err).Msg("Error when updating users")
		return
	}
	c.Lock()
	c.online.Users = users
	c.Unlock()
	l.Debug().Msgf("Found %d users", len(users))
}

func (c *SlackRTM) getUserByName(ctx context.Context, name string) (user slack.User) {
	c.RLock()
	for i := range c.online.Users {
		if c.online.Users[i].Name == name {
			c.RUnlock()
			return c.online.Users[i]
		}
	}
	c.RUnlock()
	c.updateUsers(ctx)
	c.RLock()
	for i := range c.online.Users {
		if c.online.Users[i].Name == name {
			c.RUnlock()
			return c.online.Users[i]
		}
	}
	c.RUnlock()
	return
}

func (c *SlackRTM) getUserByID(ctx context.Context, id string) (user slack.User) {
	c.RLock()
	for i := range c.online.Users {
		if c.online.Users[i].ID == id {
			c.RUnlock()
			return c.online.Users[i]
		}
	}
	c.RUnlock()
	c.updateUsers(ctx)
	c.RLock()
	for i := range c.online.Users {
		if c.online.Users[i].ID == id {
			c.RUnlock()
			return c.online.Users[i]
		}
	}
	c.RUnlock()
	return
}

func (c *SlackRTM) getChannelByName(ctx context.Context, name string) (user slack.Channel) {
	c.RLock()
	for i := range c.online.Channels {
		if c.online.Channels[i].Name == name {
			c.RUnlock()
			return c.online.Channels[i]
		}
	}
	c.RUnlock()
	c.updateChannels(ctx)
	c.RLock()
	for i := range c.online.Channels {
		if c.online.Channels[i].Name == name {
			c.RUnlock()
			return c.online.Channels[i]
		}
	}
	c.RUnlock()
	return
}

func (c *SlackRTM) getChannelByID(ctx context.Context, id string) (user slack.Channel) {
	if strings.HasPrefix(id, "D") {
		return
	}
	c.RLock()
	for i := range c.online.Channels {
		if c.online.Channels[i].ID == id {
			c.RUnlock()
			return c.online.Channels[i]
		}
	}
	c.RUnlock()
	c.updateChannels(ctx)
	c.RLock()
	for i := range c.online.Channels {
		if c.online.Channels[i].ID == id {
			c.RUnlock()
			return c.online.Channels[i]
		}
	}
	c.RUnlock()
	return
}

func (c *SlackRTM) sendToChannel(ctx context.Context, channel string, message string) {
	l := logger(ctx)
	params := slack.PostMessageParameters{
		IconURL:   c.botIcon.IconURL,
		IconEmoji: c.botIcon.IconEmoji,
		LinkNames: 1,
		AsUser:    false,
	}
	_, _, _, err := c.rtm.SendMessageContext(ctx,
		channel,
		slack.MsgOptionText(message, false),
		slack.MsgOptionParse(true),
		slack.MsgOptionPost(),
		slack.MsgOptionPostMessageParameters(params),
	)
	if err != nil {
		l.Error().Err(err).Msg("Error sending Slack message")
	}
}

func (c *SlackRTM) getID() string {
	id := atomic.AddInt64(&c.id, 1)
	return fmt.Sprintf("%s-%d-%d", c.name, c.created, id)
}
