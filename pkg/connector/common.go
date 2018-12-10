package connector

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	serviceName = "connector_catalog"
	serviceType = "core"
)

func logger() zerolog.Logger {
	return log.With().
		Str("service", serviceName).
		Str("service_type", serviceType).
		Logger()
}
