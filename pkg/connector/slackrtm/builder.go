package slackrtm

import (
	"github.com/geliar/manopus/pkg/connector"
	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/log"
	"github.com/nlopes/slack"
	"github.com/rs/zerolog"
)

func init() {
	l := logger()
	l.Debug().Msg("Registering connector in the catalog")
	connector.Register(connectorName, builder)
}

type slackLogger struct {
	log zerolog.Logger
}

func (l *slackLogger) Output(_ int, msg string) error {
	log.Debug().Msg(msg)
	return nil
}

func builder(name string, config map[string]interface{}) {
	l := logger().With().Str("connector_name", name).Logger()
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
	input.Register(name, i)
}
