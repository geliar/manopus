package slackrtm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/geliar/manopus/pkg/payload"

	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/log"
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
	stop     bool
	stopped  chan struct{}
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

func (c *SlackRTM) Send(ctx context.Context, response *payload.Response) {
	l := logger(ctx)
	l.Debug().
		Str("input_name", response.Request.Input).
		Str("input_event_id", response.Request.ID).
		Msg("Received Send event")
	var chid, text string

	switch v := response.Data["data"].(type) {
	case string:
		text = v
	case []byte:
		text = string(v)
	default:
		l.Error().Msg("Unknown type of 'data' field of response")
		return
	}
	if text == "" {
		l.Debug().
			Msg("Text field is empty")
		return
	}

	chid = c.getChannelID(ctx, response)

	if chid == "" {
		l.Error().Msg("Cannot determine channel_id")
		return
	}

	var attachments []slack.Attachment
	switch v := response.Data["attachments"].(type) {
	case []interface{}:
		buf, err := json.Marshal(v)
		if err != nil {
			l.Error().
				Err(err).
				Msg("Error marshaling attachments")
		}
		err = json.Unmarshal(buf, &attachments)
		if err != nil {
			l.Error().
				Err(err).
				Msg("Error unmarshaling attachments")
		}
	}
	c.sendToChannel(ctx, chid, attachments, text)
}

func (c *SlackRTM) Stop(ctx context.Context) {
	c.Lock()
	if c.stop {
		c.Unlock()
		return
	}
	c.stop = true
	c.Unlock()
	if c.rtm != nil {
		_ = c.rtm.Disconnect()
	}
	<-c.stopped
}

func (c *SlackRTM) serve(ctx context.Context) {
	l := logger(ctx)
	go c.rtm.ManageConnection()
	for msg := range c.rtm.IncomingEvents {
		if c.stop {
			l.Info().Msg("Shutdown request received")
			defer func() { c.stopped <- struct{}{} }()
			return
		}
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
				e := &payload.Event{
					Input: c.name,
					Type:  connectorName,
					ID:    c.getID(),
					Data: map[string]interface{}{
						"user_id":           ev.User,
						"user_name":         c.getUserByID(ctx, ev.User).Name,
						"user_display_name": c.getUserByID(ctx, ev.User).Name,
						"user_real_name":    c.getUserByID(ctx, ev.User).RealName,
						"channel_id":        ev.Channel,
						"channel_name":      c.getChannelByID(ctx, ev.Channel).Name,
						"thread_ts":         ev.ThreadTimestamp,
						"message":           ev.Text,
						"mentioned":         strings.Contains(ev.Text, fmt.Sprintf("<@%s>", c.online.User.ID)),
						"direct":            strings.HasPrefix(ev.Channel, "D"),
					},
				}
				c.sendEventToHandlers(ctx, e)
			}

		case *slack.RTMError:
			l.Debug().Err(ev).Msgf("Error: %s\n", ev.Error())
		case *slack.InvalidAuthEvent:
			l.Fatal().Msgf("Invalid credentials")
			return
		case *slack.DisconnectedEvent:
			l.Debug().Msgf("Disconnected event received")
		case *slack.ChannelCreatedEvent,
			*slack.ChannelDeletedEvent,
			*slack.ChannelArchiveEvent,
			*slack.ChannelRenameEvent,
			*slack.ChannelUnarchiveEvent:
			c.updateChannels(ctx)
		case *slack.UserChangeEvent:
			c.updateUsers(ctx)
		default:
		}
	}
}

func (c *SlackRTM) sendEventToHandlers(ctx context.Context, event *payload.Event) {
	c.RLock()
	defer c.RUnlock()
	for _, h := range c.handlers {
		go func() {
			response := h(ctx, event)
			if response != nil {
				response.Data["channel_id"], _ = response.Request.Data["channel_id"].(string)
				c.Send(ctx, response)
			}
		}()
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
			user = c.online.Users[i]
			c.RUnlock()
			return
		}
	}
	c.RUnlock()
	c.updateUsers(ctx)
	c.RLock()
	for i := range c.online.Users {
		if c.online.Users[i].Name == name {
			user = c.online.Users[i]
			c.RUnlock()
			return
		}
	}
	c.RUnlock()
	return
}

func (c *SlackRTM) getUserByDisplayName(ctx context.Context, name string) (user slack.User) {
	c.RLock()
	for i := range c.online.Users {
		if c.online.Users[i].Profile.DisplayNameNormalized == name {
			user = c.online.Users[i]
			c.RUnlock()
			return
		}
	}
	c.RUnlock()
	c.updateUsers(ctx)
	c.RLock()
	for i := range c.online.Users {
		if c.online.Users[i].Profile.DisplayNameNormalized == name {
			user = c.online.Users[i]
			c.RUnlock()
			return
		}
	}
	c.RUnlock()
	return
}

func (c *SlackRTM) getUserByID(ctx context.Context, id string) (user slack.User) {
	c.RLock()
	for i := range c.online.Users {
		if c.online.Users[i].ID == id {
			user = c.online.Users[i]
			c.RUnlock()
			return
		}
	}
	c.RUnlock()
	c.updateUsers(ctx)
	c.RLock()
	for i := range c.online.Users {
		if c.online.Users[i].ID == id {
			user = c.online.Users[i]
			c.RUnlock()
			return
		}
	}
	c.RUnlock()
	return
}

func (c *SlackRTM) getChannelByName(ctx context.Context, name string) (channel slack.Channel) {
	c.RLock()
	for i := range c.online.Channels {
		if c.online.Channels[i].Name == name {
			channel = c.online.Channels[i]
			c.RUnlock()
			return
		}
	}
	c.RUnlock()
	c.updateChannels(ctx)
	c.RLock()
	for i := range c.online.Channels {
		if c.online.Channels[i].Name == name {
			channel = c.online.Channels[i]
			c.RUnlock()
			return
		}
	}
	c.RUnlock()
	return
}

func (c *SlackRTM) getChannelByID(ctx context.Context, id string) (channel slack.Channel) {
	if strings.HasPrefix(id, "D") {
		return
	}
	c.RLock()
	for i := range c.online.Channels {
		if c.online.Channels[i].ID == id {
			channel = c.online.Channels[i]
			c.RUnlock()
			return
		}
	}
	c.RUnlock()
	c.updateChannels(ctx)
	c.RLock()
	for i := range c.online.Channels {
		if c.online.Channels[i].ID == id {
			channel = c.online.Channels[i]
			c.RUnlock()
			return
		}
	}
	c.RUnlock()
	return
}

func (c *SlackRTM) openIM(ctx context.Context, userID string) (imID string) {
	log.Debug().
		Str("slack_user_id", userID).
		Msg("Openning new Slack IM channel")
	_, _, imID, _ = c.rtm.OpenIMChannelContext(ctx, userID)
	return imID
}

func (c *SlackRTM) sendToChannel(ctx context.Context, channel string, attachments []slack.Attachment, message string) {
	l := logger(ctx)
	l.Debug().
		Str("slack_channel", channel).
		Msg("Sending message to Slack channel")
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
		slack.MsgOptionAttachments(attachments...),
	)
	if err != nil {
		l.Error().Err(err).Msg("Error sending Slack message")
	}
}

func (c *SlackRTM) getChannelID(ctx context.Context, response *payload.Response) string {
	if chID, ok := response.Data["channel_id"].(string); ok && chID != "" {
		return chID
	}
	if chname, ok := response.Data["channel_name"].(string); ok && chname != "" {
		if ch := c.getChannelByName(ctx, chname); ch.ID != "" {
			return ch.ID
		}
	}
	if userID, ok := response.Data["user_id"].(string); ok && userID != "" {
		if ch := c.openIM(ctx, userID); ch != "" {
			return ch
		}
	}
	if username, ok := response.Data["user_name"].(string); ok && username != "" {
		if user := c.getUserByName(ctx, username); user.ID != "" {
			if ch := c.openIM(ctx, user.ID); ch != "" {
				return ch
			}
		}
	}
	if username, ok := response.Data["user_display_name"].(string); ok && username != "" {
		if user := c.getUserByDisplayName(ctx, username); user.ID != "" {
			if ch := c.openIM(ctx, user.ID); ch != "" {
				return ch
			}
		}
	}
	return ""
}

func (c *SlackRTM) getID() string {
	id := atomic.AddInt64(&c.id, 1)
	return fmt.Sprintf("%s-%d-%d", c.name, c.created, id)
}
