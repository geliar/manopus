package bitbucket

import (
	"context"
	"net/http"
	"time"

	"github.com/geliar/manopus/pkg/connector"
	mhttp "github.com/geliar/manopus/pkg/http"
	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/output"

	cbitbucket "github.com/ktrysmt/go-bitbucket"
	whbitbucket "gopkg.in/go-playground/webhooks.v5/bitbucket"
)

func init() {
	ctx := log.Logger.WithContext(context.Background())
	connector.Register(ctx, serviceName, builder)
}

func builder(ctx context.Context, name string, config map[string]interface{}) {
	l := logger(ctx)
	l = l.With().Str("connector_name", name).Logger()
	ctx = l.WithContext(ctx)
	l.Debug().Msgf("Initializing new instance of %s", serviceName)
	i := new(GitHub)
	i.created = time.Now().UTC().UnixNano()
	i.name = name
	i.stopCh = make(chan struct{})
	if config != nil {
		uuid, _ := config["webhook_uuid"].(string)
		callback, _ := config["webhook_callback"].(string)
		if callback != "" {
			var err error
			i.hook, err = whbitbucket.New(whbitbucket.Options.UUID(uuid))
			if err != nil {
				l.Error().Err(err).Msg("Error applying Bitbucket UUID")
			}
			mhttp.AddHandler(ctx, callback, http.HandlerFunc(i.WebhookHandler))
			l.Info().Msgf("Bitbucket webhook on path %s", callback)
		}
		okey, _ := config["oauth2_key"].(string)
		osecret, _ := config["oauth2_secret"].(string)
		if okey != "" && osecret != "" {
			i.client = cbitbucket.NewOAuthClientCredentials(okey, osecret)
		}
	}
	input.Register(ctx, name, i)
	output.Register(ctx, name, i)
}
