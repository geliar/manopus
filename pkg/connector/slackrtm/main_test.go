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
	t.Run("Connect", func(t *testing.T) {
		i.created = time.Now().UnixNano()

		i.name = "test"
		i.debug = false
		i.token = os.Getenv("SLACK_TOKEN")

		i.channels = []string{os.Getenv("SLACK_CHANNEL")}

		a.NoError(i.validate())

		client := slack.New(i.token)
		slack.SetLogger(&slackLogger{log: l})
		client.SetDebug(i.debug)
		t.Log("Starting RTM")
		i.rtm = client.NewRTM()
		go i.serve(ctx)
		for n := 0; n < 20; n++ {
			runtime.Gosched()
			if i.online.User.ID != "" && len(i.online.Channels) != 0 {
				break
			}
			t.Log("Waiting for RTM to start")
			time.Sleep(time.Millisecond * 500)
		}
		a.NotEmpty(i.online.User.ID)
	})
	t.Run("getUser", func(t *testing.T) {
		i.online.Users = i.online.Users[:0]
		a.Equal(i.online.User.ID, i.getUserByName(ctx, i.online.User.Name).ID)
		a.Equal(i.online.User.Name, i.getUserByID(ctx, i.online.User.ID).Name)
		//From cache
		a.Equal(i.online.User.ID, i.getUserByName(ctx, i.online.User.Name).ID)
		a.Equal(i.online.User.Name, i.getUserByID(ctx, i.online.User.ID).Name)
	})
	t.Run("getChannel", func(t *testing.T) {
		ch := i.online.Channels[0]
		i.online.Channels = i.online.Channels[:0]
		a.Equal(ch.ID, i.getChannelByName(ctx, ch.Name).ID)
		println("111")
		a.Equal(ch.Name, i.getChannelByID(ctx, ch.ID).Name)
		//From cache
		a.Equal(ch.ID, i.getChannelByName(ctx, ch.Name).ID)
		a.Equal(ch.Name, i.getChannelByID(ctx, ch.ID).Name)
	})
	//_ = i.rtm.Disconnect()
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
