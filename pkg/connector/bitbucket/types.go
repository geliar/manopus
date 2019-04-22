package bitbucket

import (
	whbitbucket "gopkg.in/go-playground/webhooks.v5/bitbucket"
)

const requestTypePullRequestCreated = string(whbitbucket.PullRequestCreatedEvent)

type requestPullRequestCreated struct {
	whbitbucket.PullRequestCreatedPayload
}

const requestTypePullRequestApproved = string(whbitbucket.PullRequestApprovedEvent)

type requestPullRequestApproved struct {
	whbitbucket.PullRequestApprovedPayload
}

const requestTypeRepoPush = string(whbitbucket.RepoPushEvent)

type requestPush struct {
	whbitbucket.RepoPushPayload
}
