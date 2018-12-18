package sequencer

import (
	"github.com/geliar/manopus/pkg/log"
	"github.com/rs/zerolog"
)

const (
	serviceName = "sequencer"
	serviceType = "core"
)

func logger() zerolog.Logger {
	return log.With().
		Str("service", serviceName).
		Str("service_type", serviceType).
		Logger()
}
