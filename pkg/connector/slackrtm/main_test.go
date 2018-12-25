// +build integration

package slackrtm

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"

	"github.com/geliar/manopus/pkg/log"
)

func TestSlack(t *testing.T) {
	a := assert.New(t)
	l := log.Logger
	ctx := l.WithContext(context.Background())
	i := new(SlackRTM)
	i.created = time.Now().UnixNano()
	i.name = "test"
	i.debug = false
	i.token = os.Getenv("SLACK_TOKEN")

	i.channels = []string{os.Getenv("SLACK_CHANNEL")}

	a.NoError(i.validate())

	client := slack.New(i.token)
	slack.SetLogger(&slackLogger{log: l})
	client.SetDebug(i.debug)
	i.rtm = client.NewRTM()
	go i.serve(ctx)
	for n := 0; n < 10; n++ {
		if i.online.User.ID != "" {
			break
		}
		time.Sleep(time.Millisecond * 500)
	}
	runtime.Gosched()
	//time.Sleep(time.Second * 2)
	a.NotEmpty(i.online.User.ID)
	a.Equal(i.online.User.ID, i.getUserByName(ctx, i.online.User.Name).ID)
	_ = i.rtm.Disconnect()
}

func TestSlackBuilder(t *testing.T) {
	l := log.Logger
	ctx := l.WithContext(context.Background())
	config := map[string]interface{}{
		"debug": true,
		"token": os.Getenv("SLACK_TOKEN"),
		"channels": []interface{}{
			os.Getenv("SLACK_CHANNEL"),
		},
		"bot_icon_url":   "url",
		"bot_icon_emoji": "emoji",
	}
	builder(ctx, "testslack", config)
}
