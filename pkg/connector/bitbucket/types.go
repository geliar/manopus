package bitbucket

import (
	whbitbucket "gopkg.in/go-playground/webhooks.v5/bitbucket"
)

const RequestTypePullRequestCreated = string(whbitbucket.PullRequestCreatedEvent)

type RequestPullRequestCreated struct {
	whbitbucket.PullRequestCreatedPayload
}

const RequestTypePullRequestApproved = string(whbitbucket.PullRequestApprovedEvent)

type RequestPullRequestApproved struct {
	whbitbucket.PullRequestApprovedPayload
}

const RequestTypeRepoPush = string(whbitbucket.RepoPushEvent)

type RequestPush struct {
	whbitbucket.RepoPushPayload
}
