package slackrtm

import (
	"context"

	"github.com/geliar/manopus/pkg/connector"
	"github.com/geliar/manopus/pkg/input"
	"github.com/nlopes/slack"
	"github.com/rs/zerolog"
)

func init() {
	ctx := context.Background()
	l := logger(ctx)
	l.Debug().Msg("Registering connector in the catalog")
	connector.Register(l.WithContext(ctx), connectorName, builder)
}

type slackLogger struct {
	log zerolog.Logger
}

func (s *slackLogger) Output(_ int, msg string) error {
	l := logger(context.Background())
	l.Debug().Msg(msg)
	return nil
}

func builder(ctx context.Context, name string, config map[string]interface{}) {
	l := logger(ctx).With().Str("connector_name", name).Logger()
	l.Debug().Msg("Registering input in the registry")
	i := new(SlackRTM)
	i.debug, _ = config["debug"].(bool)
	i.token, _ = config["token"].(string)
	i.channels, _ = config["token"].([]string)
	i.botName, _ = config["botName"].(string)
	i.botIcon, _ = config["botIcon"].(string)
	if i.validate() != nil {
		l.Fatal().Msg("Cannot validate parameters of connector")
	}

	i.client = slack.New(i.token)
	slack.SetLogger(&slackLogger{log: l})
	i.client.SetDebug(i.debug)
	i.rtm = i.client.NewRTM()
	input.Register(ctx, name, i)
}
