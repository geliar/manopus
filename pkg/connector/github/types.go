package github

import (
	whgithub "gopkg.in/go-playground/webhooks.v5/github"
)

const requestTypePullRequest = string(whgithub.PullRequestEvent)

type requestPullRequest struct {
	whgithub.PullRequestPayload
	Issue gitHubIssue `json:"issue"`
}

const requestTypeIssueComment = string(whgithub.IssueCommentEvent)

type requestIssueComment struct {
	whgithub.IssueCommentPayload
	gitHubPullRequest
}

const requestTypePush = string(whgithub.PushEvent)

type requestPush struct {
	whgithub.PushPayload
	Branch string     `json:"branch"`
	Head   gitHubHead `json:"head"`
}

type gitHubHead struct {
	Ref  string `json:"ref"`
	Sha  string `json:"sha"`
	User struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
	} `json:"user"`
}

type gitHubIssue struct {
	URL         string `json:"url"`
	LabelsURL   string `json:"labels_url"`
	CommentsURL string `json:"comments_url"`
	EventsURL   string `json:"events_url"`
	HTMLURL     string `json:"html_url"`
	ID          int64  `json:"id"`
	Number      int64  `json:"number"`
	Title       string `json:"title"`
}

type gitHubPullRequest struct {
	URL      string `json:"url"`
	ID       int64  `json:"id"`
	HTMLURL  string `json:"html_url"`
	DiffURL  string `json:"diff_url"`
	PatchURL string `json:"patch_url"`
	IssueURL string `json:"issue_url"`
	Number   int64  `json:"number"`
	State    string `json:"state"`
	Locked   bool   `json:"locked"`
	Title    string `json:"title"`
}
