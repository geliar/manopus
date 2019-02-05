package timer

import (
	"context"
	"time"

	"github.com/geliar/manopus/pkg/connector"
	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/output"
)

func init() {
	ctx := log.Logger.WithContext(context.Background())
	connector.Register(ctx, connectorName, builder)
}

func builder(ctx context.Context, name string, config map[string]interface{}) {
	l := logger(ctx)
	l = l.With().Str("connector_name", name).Logger()
	ctx = l.WithContext(ctx)
	l.Debug().Msgf("Initializing new instance of %s", connectorName)
	i := new(Timer)
	i.created = time.Now().UnixNano()
	i.name = name
	input.Register(ctx, name, i)
	output.Register(ctx, name, i)
}
