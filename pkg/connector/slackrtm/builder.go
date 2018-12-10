package slackrtm

import (
	"github.com/DLag/manopus/pkg/connector"
	"github.com/DLag/manopus/pkg/input"
	"github.com/DLag/manopus/pkg/log"
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
	l := logger().With().Str("connector", name).Logger()
	l.Debug().Msg("Registering input in the registry")
	i := new(SlackRTM)
	i.debug, _ = config["debug"].(bool)
	i.token, _ = config["token"].(string)
	i.channels, _ = config["token"].([]string)
	i.botName, _ = config["botName"].(string)
	i.botIcon, _ = config["botIcon"].(string)

	i.client = slack.New(i.token)
	slack.SetLogger(&slackLogger{log: l})
	i.client.SetDebug(i.debug)

	input.Register(name, i)
}
