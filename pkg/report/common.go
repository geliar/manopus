package report

import (
	"context"

	"github.com/geliar/manopus/pkg/log"

	"github.com/rs/zerolog"
)

const (
	serviceName = "report"
	serviceType = "core"
)

func logger(ctx context.Context) zerolog.Logger {
	return log.Ctx(ctx).With().
		Str("service", serviceName).
		Str("service_type", serviceType).
		Logger()
}
