package http

import (
	"context"
	"time"

	"github.com/geliar/manopus/pkg/connector"
	"github.com/geliar/manopus/pkg/http"
	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/output"
)

func init() {
	connector.Register(log.Logger.WithContext(context.Background()), connectorName, builder)
}

func builder(ctx context.Context, name string, config map[string]interface{}) {
	l := logger(ctx)
	l = l.With().Str("connector_name", name).Logger()
	ctx = l.WithContext(ctx)
	l.Debug().Msgf("Initializing new instance of %s", connectorName)
	i := new(HTTP)
	i.created = time.Now().UTC().UnixNano()
	i.name = name
	if i.validate() != nil {
		l.Fatal().Msg("Cannot validate parameters of connector")
	}
	input.Register(ctx, name, i)
	output.Register(ctx, name, i)
	http.SetDefaultHandler(ctx, i)
}
