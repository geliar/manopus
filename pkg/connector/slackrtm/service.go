package slackrtm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/davecgh/go-spew/spew"

	"github.com/geliar/manopus/pkg/input"
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
	c.Lock()
	defer c.Unlock()
	l := logger(ctx)
	l.Info().Msg("Registered new handler")
	c.handlers = append(c.handlers, handler)
}

func (*SlackRTM) SendEvent(input input.Event) {
	panic("implement me")
}

func (*SlackRTM) Stop(ctx context.Context) {
	panic("implement me")
}

func (c *SlackRTM) serve(ctx context.Context) {
	l := logger(ctx)
	go c.rtm.ManageConnection()
	for msg := range c.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

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
				e := input.Event{
					Input: c.name,
					Type:  connectorName,
					ID:    c.getID(),
					Data: map[string]interface{}{
						"user": map[string]interface{}{
							"id":   ev.User,
							"name": c.getUserByID(ctx, ev.User).Name,
						},
						"channel": map[string]interface{}{
							"id":   ev.Channel,
							"name": c.getChannelByID(ctx, ev.Channel).Name,
						},
						"message":   ev.Text,
						"mentioned": strings.Contains(ev.Text, fmt.Sprintf("<@%s>", c.online.User.ID)),
					},
				}
				spew.Dump(e)
				c.sendToChannels(ctx, "Hello <@"+ev.User+">")
			}

		case *slack.RTMError:
			l.Debug().Err(ev).Msgf("Error: %s\n", ev.Error())
		case *slack.InvalidAuthEvent:
			l.Debug().Msgf("Invalid credentials")
			_ = c.rtm.Disconnect()
			return

		default:

			// Ignore other events..
			// fmt.Printf("Unexpected: %v\n", msg.Data)
		}
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
			return c.online.Users[i]
		}
	}
	c.RUnlock()
	c.updateUsers(ctx)
	c.RLock()
	for i := range c.online.Users {
		if c.online.Users[i].Name == name {
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
			return c.online.Users[i]
		}
	}
	c.RUnlock()
	c.updateUsers(ctx)
	c.RLock()
	for i := range c.online.Users {
		if c.online.Users[i].ID == id {
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
			return c.online.Channels[i]
		}
	}
	c.RUnlock()
	c.updateChannels(ctx)
	c.RLock()
	for i := range c.online.Channels {
		if c.online.Channels[i].Name == name {
			return c.online.Channels[i]
		}
	}
	c.RUnlock()
	return
}

func (c *SlackRTM) getChannelByID(ctx context.Context, id string) (user slack.Channel) {
	c.RLock()
	for i := range c.online.Channels {
		if c.online.Channels[i].ID == id {
			return c.online.Channels[i]
		}
	}
	c.RUnlock()
	c.updateChannels(ctx)
	c.RLock()
	for i := range c.online.Channels {
		if c.online.Channels[i].ID == id {
			return c.online.Channels[i]
		}
	}
	c.RUnlock()
	return
}

func (c *SlackRTM) sendToChannels(ctx context.Context, message string) {
	l := logger(ctx)
	params := slack.PostMessageParameters{
		IconURL:   c.botIcon.IconURL,
		IconEmoji: c.botIcon.IconEmoji,
		LinkNames: 1,
		AsUser:    false,
	}
	for _, ch := range c.channels {
		_, _, _, err := c.rtm.SendMessageContext(ctx,
			c.getChannelByName(ctx, ch).ID,
			slack.MsgOptionText(message, false),
			slack.MsgOptionParse(true),
			slack.MsgOptionPost(),
			slack.MsgOptionPostMessageParameters(params),
		)
		if err != nil {
			l.Error().Err(err).Msg("Error sending Slack message")
		}
	}
}

func (c *SlackRTM) getID() string {
	id := atomic.AddInt64(&c.id, 1)
	return fmt.Sprintf("%s-%d-%d", c.name, c.created, id)
}
