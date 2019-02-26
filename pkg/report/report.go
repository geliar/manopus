package report

import (
	"context"
)

var config Config

func Init(ctx context.Context, cfg Config) {
	l := logger(ctx)
	config = cfg
	l.Info().Str("report_name", config.Driver).Msgf("Configured reporter %s", config.Driver)
	if _, ok := catalog.builders[config.Driver]; !ok {
		l.Fatal().
			Str("report_name", config.Driver).
			Msgf("Cannot find driver with name '%s'", config.Driver)
	}
}
