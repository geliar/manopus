package slackrtm

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/geliar/manopus/pkg/log"
)

func TestSlackBuilder(t *testing.T) {
	l := log.Output(ioutil.Discard)
	ctx := l.WithContext(context.Background())
	config := map[string]interface{}{
		"debug": true,
		"token": "some_token",
		"channels": []interface{}{
			"1",
			"2",
		},
		"bot_icon_url":   "url",
		"bot_icon_emoji": "emoji",
	}
	builder(ctx, "testslack", config)

}
