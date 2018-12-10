package input

import (
	"github.com/DLag/manopus/pkg/log"
	"github.com/rs/zerolog"
)

const (
	serviceName = "input_catalog"
	serviceType = "core"
)

func logger() zerolog.Logger {
	return log.With().
		Str("service", serviceName).
		Str("service_type", serviceType).
		Logger()
}
