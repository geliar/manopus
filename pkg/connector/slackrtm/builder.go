package slackrtm

import (
	"context"
	"time"

	"github.com/geliar/manopus/pkg/connector"
	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/output"

	"github.com/nlopes/slack"
	"github.com/rs/zerolog"
)

func init() {
	ctx := log.Logger.WithContext(context.Background())
	l := logger(ctx)
	l.Debug().Msg("Registering connector in the catalog")
	connector.Register(l.WithContext(ctx), connectorName, builder)
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
	l.Debug().Msg("Registering input in the registry")
	i := new(SlackRTM)
	i.created = time.Now().UnixNano()
	i.name = name
	i.debug, _ = config["debug"].(bool)
	i.token, _ = config["token"].(string)
	channels, _ := config["channels"].([]interface{})
	for _, ch := range channels {
		if chs, ok := ch.(string); ok {
			i.channels = append(i.channels, chs)
		}
	}
	iconURL, _ := config["bot_icon_url"].(string)
	iconEmoji, _ := config["bot_icon_emoji"].(string)
	i.botIcon.IconURL = iconURL
	i.botIcon.IconEmoji = iconEmoji
	if i.validate() != nil {
		l.Fatal().Msg("Cannot validate parameters of connector")
	}

	client := slack.New(i.token)
	slack.SetLogger(&slackLogger{log: l})
	client.SetDebug(i.debug)
	i.rtm = client.NewRTM()
	go i.serve(ctx)
	input.Register(ctx, name, i)
	output.Register(ctx, name, i)
}
