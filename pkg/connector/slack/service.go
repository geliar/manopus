package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
	"github.com/rs/zerolog/hlog"

	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/payload"
)

type SlackRTM struct {
	created int64
	id      int64
	name    string
	config  struct {
		debug             bool
		token             string
		verificationToken string
		messageTypes      map[string]struct{}
		botIcon           slack.Icon
		rtm               bool
	}
	online struct {
		Channels []slack.Channel
		Users    []slack.User
		User     slack.UserDetails
	}
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
	c.Lock()
	defer c.Unlock()
	c.handlers = append(c.handlers, handler)
}

func (c *SlackRTM) Send(ctx context.Context, response *payload.Response) map[string]interface{} {
	l := logger(ctx)
	l.Debug().
		Str("input_name", response.Request.Input).
		Str("input_event_id", response.Request.ID).
		Msg("Received Send event")
	var text string
	switch v := response.Data["data"].(type) {
	case string:
		text = v
	case []byte:
		text = string(v)
	default:
		l.Error().Msgf("Unknown type of 'data' field of response %T", response.Data["data"])
		return nil
	}
	if text == "" {
		l.Debug().
			Msg("Text field is empty")
		return nil
	}

	chids := c.getChannelIDs(ctx, response)

	if chids == nil {
		l.Warn().Msg("Cannot determine channel_id")
		return nil
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
				Msg("Error unmarshalling attachments")
		}
	}
	res := c.sendToChannels(ctx, chids, attachments, text)
	if res == nil {
		return nil
	}
	return map[string]interface{}{"result": res}
}

func (c *SlackRTM) Stop(ctx context.Context) {
	c.Lock()
	if c.stop {
		c.Unlock()
		return
	}
	c.stop = true
	c.Unlock()
	if c.config.rtm && c.rtm != nil {
		_ = c.rtm.Disconnect()
		<-c.stopped
	}
}

func (c *SlackRTM) InteractionCallbackHandler(w http.ResponseWriter, r *http.Request) {
	l := hlog.FromRequest(r)
	ctx := l.WithContext(context.Background())
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		l.Error().Err(err).Msg("Cannot read HTTP body")
		return
	}
	body := buf.String()
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		l.Error().Msg("Unrecognized type of payload")
		return
	}
	val, err := url.ParseQuery(body)
	if err != nil {
		l.Error().Err(err).Msg("Cannot parse application/x-www-form-urlencoded event")
		return
	}
	body = val.Get("payload")
	l.Debug().Str("content-type", r.Header.Get("Content-Type")).Str("body", body).Msg("Interactive event")
	var ev slack.InteractionCallback
	err = json.Unmarshal([]byte(body), &ev)
	if err != nil {
		l.Error().Err(err).Msg("Cannot parse Slack event payload")
		return
	}
	if ev.Token != c.config.verificationToken {
		l.Error().Msg("Wrong verification token")
		return
	}
	var actions []MessageAction
	for i := range ev.Actions {
		actions = append(actions, MessageAction{
			Name:  ev.Actions[i].Name,
			Value: ev.Actions[i].Value,
			Text:  ev.Actions[i].Text,
		})
	}

	data := RequestInteraction{
		SlackMessage: SlackMessage{
			UserID:          ev.User.ID,
			UserName:        c.getUserByID(ctx, ev.User.ID).Name,
			UserDisplayName: c.getUserByID(ctx, ev.User.ID).Name,
			UserRealName:    c.getUserByID(ctx, ev.User.ID).RealName,
			ChannelID:       ev.Channel.ID,
			ChannelName:     c.getChannelByID(ctx, ev.Channel.ID).Name,
			ThreadTS:        ev.Message.ThreadTimestamp,
			Timestamp:       ev.Message.Timestamp,
			Message:         ev.Message.Text,
			Direct:          strings.HasPrefix(ev.Channel.ID, "D"),
		},
		CallbackID: ev.CallbackID,
		Actions:    actions,
	}

	e := &payload.Event{
		Input: c.name,
		Type:  RequestTypeInteraction,
		ID:    c.getID(),
		Data:  data,
	}

	c.sendEventToHandlers(ctx, ev.Channel.ID, e)
}

func (c *SlackRTM) EventCallbackHandler(w http.ResponseWriter, r *http.Request) {
	l := hlog.FromRequest(r)
	ctx := l.WithContext(context.Background())
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		l.Error().Err(err).Msg("Cannot read HTTP body")
		return
	}
	l.Debug().Str("content-type", r.Header.Get("Content-Type")).Str("body", buf.String()).Msg("Slack event")
	body := buf.String()
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: c.config.verificationToken}))
	if err != nil {
		l.Error().Err(err).Msg("Cannot parse event")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"challenge":"` + r.Challenge + `"}`))
		return
	case slackevents.CallbackEvent:
		eventData := eventsAPIEvent.Data.(*slackevents.EventsAPICallbackEvent)
		innerEvent := eventsAPIEvent.InnerEvent
		e := new(payload.Event)
		var channel string
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			if ev.User == "" {
				l.Debug().Msg("Message from bot itself, skipping")
				return
			}
			channel = ev.Channel
			data := RequestMessage{
				SlackMessage: SlackMessage{
					UserID:          ev.User,
					UserName:        c.getUserByID(ctx, ev.User).Name,
					UserDisplayName: c.getUserByID(ctx, ev.User).Name,
					UserRealName:    c.getUserByID(ctx, ev.User).RealName,
					ChannelID:       ev.Channel,
					ChannelName:     c.getChannelByID(ctx, ev.Channel).Name,
					ThreadTS:        ev.ThreadTimeStamp,
					Timestamp:       ev.TimeStamp,
					Message:         ev.Text,
					Direct:          strings.HasPrefix(ev.Channel, "D"),
				},
				Mentioned: true,
			}
			e = &payload.Event{
				Input: c.name,
				Type:  RequestTypeMessage,
				ID:    c.getID(),
				Data:  data,
			}
		case *slackevents.MessageEvent:
			for _, u := range eventData.AuthedUsers {
				if ev.User == "" || ev.User == u {
					l.Debug().Msg("Message from bot itself, skipping")
					return
				}
			}
			channel = ev.Channel
			data := RequestMessage{
				SlackMessage: SlackMessage{
					UserID:          ev.User,
					UserName:        c.getUserByID(ctx, ev.User).Name,
					UserDisplayName: c.getUserByID(ctx, ev.User).Name,
					UserRealName:    c.getUserByID(ctx, ev.User).RealName,
					ChannelID:       ev.Channel,
					ChannelName:     c.getChannelByID(ctx, ev.Channel).Name,
					ThreadTS:        ev.ThreadTimeStamp,
					Timestamp:       ev.TimeStamp,
					Message:         ev.Text,
					Direct:          strings.HasPrefix(ev.Channel, "D"),
				},
			}
			for _, u := range eventData.AuthedUsers {
				if strings.Contains(ev.Text, fmt.Sprintf("<@%s>", u)) {
					data.Mentioned = true
				}
			}
			e = &payload.Event{
				Input: c.name,
				Type:  RequestTypeMessage,
				ID:    c.getID(),
				Data:  data,
			}
		}
		c.sendEventToHandlers(ctx, channel, e)
	}
}

func (c *SlackRTM) rtmServe(ctx context.Context) {
	l := logger(ctx)
	go c.rtm.ManageConnection()
	l.Debug().Msg("Slack RTM connection has been started")
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
			if _, ok := c.config.messageTypes[ev.SubType]; ev.User != "" && (ev.SubType == "" || ok) {
				data := RequestMessage{
					SlackMessage: SlackMessage{
						UserID:          ev.User,
						UserName:        c.getUserByID(ctx, ev.User).Name,
						UserDisplayName: c.getUserByID(ctx, ev.User).Name,
						UserRealName:    c.getUserByID(ctx, ev.User).RealName,
						ChannelID:       ev.Channel,
						ChannelName:     c.getChannelByID(ctx, ev.Channel).Name,
						ThreadTS:        ev.ThreadTimestamp,
						Timestamp:       ev.Timestamp,
						Message:         ev.Text,
						Direct:          strings.HasPrefix(ev.Channel, "D"),
					},
					Mentioned: strings.Contains(ev.Text, fmt.Sprintf("<@%s>", c.online.User.ID)),
				}
				e := &payload.Event{
					Input: c.name,
					Type:  RequestTypeMessage,
					ID:    c.getID(),
					Data:  data,
				}
				c.sendEventToHandlers(ctx, ev.Channel, e)
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

func (c *SlackRTM) sendEventToHandlers(ctx context.Context, channel string, event *payload.Event) {
	c.RLock()
	defer c.RUnlock()
	for _, h := range c.handlers {
		go func(handler input.Handler) {
			res := handler(ctx, event)
			if res != nil {
				response := new(payload.Response)
				response.Request = event
				response.Data = make(map[string]interface{})
				response.Data = map[string]interface{}{
					"data":       res,
					"channel_id": channel,
				}
				response.ID = event.ID
				response.Request = event
				response.Output = serviceName
				c.Send(ctx, response)
			}
		}(h)
	}
}

func (c *SlackRTM) updateChannels(ctx context.Context) {
	l := logger(ctx)
	l.Debug().Msg("Updating Slack messageTypes")
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
			l.Error().Err(err).Msg("Error when updating messageTypes")
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
	l.Debug().Msgf("Found %d messageTypes", len(resChannels))
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

func (c *SlackRTM) sendToChannels(ctx context.Context, channels []string, attachments []slack.Attachment, message string) (result []map[string]interface{}) {
	l := logger(ctx)
	l.Debug().
		Strs("slack_channel", channels).
		Msg("Sending message to Slack channels")
	params := slack.PostMessageParameters{
		IconURL:   c.config.botIcon.IconURL,
		IconEmoji: c.config.botIcon.IconEmoji,
		LinkNames: 1,
		AsUser:    false,
		Markdown:  true,
	}

	for _, channel := range channels {
		var err error
		r := make(map[string]interface{})
		r["channel"], r["thread_ts"], r["text"], err = c.rtm.SendMessageContext(ctx,
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
		result = append(result, r)
	}
	return
}

func (c *SlackRTM) getChannelIDs(ctx context.Context, response *payload.Response) []string {
	switch v := response.Data["channel_id"].(type) {
	case string:
		return []string{v}
	case []interface{}:
		var res []string
		for i := range v {
			if s, ok := v[i].(string); ok {
				res = append(res, s)
			}
		}
		return res
	}
	switch v := response.Data["channel_name"].(type) {
	case string:
		if ch := c.getChannelByName(ctx, v); ch.ID != "" {
			return []string{ch.ID}
		}
	case []interface{}:
		var res []string
		for i := range v {
			if s, ok := v[i].(string); ok {
				if ch := c.getChannelByName(ctx, s); ch.ID != "" {
					res = append(res, ch.ID)
				}
			}
		}
		return res
	}

	switch v := response.Data["user_id"].(type) {
	case string:
		if ch := c.openIM(ctx, v); ch != "" {
			return []string{ch}
		}
	case []interface{}:
		var res []string
		for i := range v {
			if s, ok := v[i].(string); ok {
				if ch := c.openIM(ctx, s); ch != "" {
					res = append(res, ch)
				}
			}
		}
		return res
	}

	switch v := response.Data["user_name"].(type) {
	case string:
		if user := c.getUserByName(ctx, v); user.ID != "" {
			if ch := c.openIM(ctx, user.ID); ch != "" {
				return []string{ch}
			}
		}
	case []interface{}:
		var res []string
		for i := range v {
			if s, ok := v[i].(string); ok {
				if user := c.getUserByName(ctx, s); user.ID != "" {
					if ch := c.openIM(ctx, user.ID); ch != "" {
						res = append(res, ch)
					}
				}
			}
		}
		return res
	}

	switch v := response.Data["user_display_name"].(type) {
	case string:
		if user := c.getUserByDisplayName(ctx, v); user.ID != "" {
			if ch := c.openIM(ctx, user.ID); ch != "" {
				return []string{ch}
			}
		}
	case []interface{}:
		var res []string
		for i := range v {
			if s, ok := v[i].(string); ok {
				if user := c.getUserByDisplayName(ctx, s); user.ID != "" {
					if ch := c.openIM(ctx, user.ID); ch != "" {
						res = append(res, ch)
					}
				}
			}
		}
		return res
	}

	return nil
}

func (c *SlackRTM) getID() string {
	id := atomic.AddInt64(&c.id, 1)
	return fmt.Sprintf("%s-%d-%d", c.name, c.created, id)
}
