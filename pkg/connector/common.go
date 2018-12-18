package connector

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	serviceName = "connector_catalog"
	serviceType = "core"
)

func logger(ctx context.Context) zerolog.Logger {
	return log.Ctx(ctx).With().
		Str("service", serviceName).
		Str("service_type", serviceType).
		Logger()
}
