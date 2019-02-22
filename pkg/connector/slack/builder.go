package slack

import (
	"context"
	"net/http"
	"time"

	mhttp "github.com/geliar/manopus/pkg/http"

	"github.com/geliar/manopus/pkg/connector"
	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/output"

	"github.com/nlopes/slack"
	"github.com/rs/zerolog"
)

func init() {
	ctx := log.Logger.WithContext(context.Background())
	connector.Register(ctx, connectorName, builder)
}

type slackLogger struct {
	log zerolog.Logger
}

func (s *slackLogger) Output(_ int, msg string) error {
	s.log.Debug().Msg(msg)
	return nil
}

func builder(ctx context.Context, name string, config map[string]interface{}) {
	l := logger(ctx)
	l = l.With().Str("connector_name", name).Logger()
	ctx = l.WithContext(ctx)
	l.Debug().Msgf("Initializing new instance of %s", connectorName)
	i := new(SlackRTM)
	i.created = time.Now().UTC().UnixNano()
	i.name = name
	i.config.debug, _ = config["debug"].(bool)
	i.config.rtm, _ = config["rtm"].(bool)
	i.config.token, _ = config["token"].(string)
	i.config.verificationToken, _ = config["verification_token"].(string)
	messageTypes, _ := config["message_types"].([]interface{})
	i.config.messageTypes = map[string]struct{}{}
	for _, mt := range messageTypes {
		if mts, ok := mt.(string); ok {
			i.config.messageTypes[mts] = struct{}{}
		}
	}
	iconURL, _ := config["bot_icon_url"].(string)
	iconEmoji, _ := config["bot_icon_emoji"].(string)
	i.config.botIcon.IconURL = iconURL
	i.config.botIcon.IconEmoji = iconEmoji
	if i.validate() != nil {
		l.Fatal().Msg("Cannot validate parameters of connector")
	}
	i.stopped = make(chan struct{})
	client := slack.New(i.config.token, slack.OptionDebug(i.config.debug), slack.OptionLog(&slackLogger{log: l}))
	i.rtm = client.NewRTM()
	if i.config.rtm {
		go i.rtmServe(ctx)
	}
	input.Register(ctx, name, i)
	output.Register(ctx, name, i)

	if eventCallback, _ := config["event_callback"].(string); eventCallback != "" {
		mhttp.AddHandler(ctx, eventCallback, http.HandlerFunc(i.EventCallbackHandler))
	}
	if interactionCallback, _ := config["interaction_callback"].(string); interactionCallback != "" {
		mhttp.AddHandler(ctx, interactionCallback, http.HandlerFunc(i.InteractionCallbackHandler))
	}
}
