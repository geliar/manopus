package github

import (
	"context"
	"net/http"
	"time"

	"github.com/geliar/manopus/pkg/connector"
	mhttp "github.com/geliar/manopus/pkg/http"
	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/output"

	cgithub "github.com/google/go-github/v24/github"
	whgithub "gopkg.in/go-playground/webhooks.v5/github"

	"golang.org/x/oauth2"
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
		secret, _ := config["webhook_secret"].(string)
		callback, _ := config["webhook_callback"].(string)
		if secret != "" && callback != "" {
			var err error
			i.hook, err = whgithub.New(whgithub.Options.Secret(secret))
			if err != nil {
				l.Error().Err(err).Msg("Error applying GitHub secret")
			}
			mhttp.AddHandler(ctx, callback, http.HandlerFunc(i.webhookHandler))
			l.Info().Msgf("GitHub webhook on path %s", callback)
		}
		token, _ := config["oauth2_token"].(string)
		if token != "" {
			ts := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			)
			tc := oauth2.NewClient(ctx, ts)
			i.client = cgithub.NewClient(tc)
		}
	}
	input.Register(ctx, name, i)
	output.Register(ctx, name, i)
}
