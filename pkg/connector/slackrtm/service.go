package slackrtm

import (
	"context"
	"github.com/geliar/manopus/pkg/input"
	"github.com/nlopes/slack"
)

type SlackRTM struct {
	name string
	debug bool
	token string
	channels []string
	botName string
	botIcon string
	client *slack.Client
	rtm *slack.RTM
}

func (*SlackRTM) validate() error {
	return nil
}

func (c *SlackRTM) Name() string {
	return c.name
}

func (*SlackRTM) RegisterHandler(handler input.Handler) {
	panic("implement me")
}

func (*SlackRTM) SendEvent(handler input.Handler) {
	panic("implement me")
}

func (*SlackRTM) Stop(ctx context.Context) {
	panic("implement me")
}


