package github

import (
	"gopkg.in/go-playground/webhooks.v5/github"
	whgithub "gopkg.in/go-playground/webhooks.v5/github"
)

const RequestTypePullRequest = string(github.PullRequestEvent)

type RequestPullRequest struct {
	github.PullRequestPayload
	Issue GitHubIssue `json:"issue"`
}

const RequestTypeIssueComment = string(github.IssueCommentEvent)

type RequestIssueComment struct {
	github.IssueCommentPayload
	GitHubPullRequest
}

const RequestTypePush = string(whgithub.PushEvent)

type RequestPush struct {
	github.PushPayload
	Branch string     `json:"branch"`
	Head   GitHubHead `json:"head"`
}

type GitHubHead struct {
	Ref  string `json:"ref"`
	Sha  string `json:"sha"`
	User struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
	} `json:"user"`
}

type GitHubIssue struct {
	URL         string `json:"url"`
	LabelsURL   string `json:"labels_url"`
	CommentsURL string `json:"comments_url"`
	EventsURL   string `json:"events_url"`
	HTMLURL     string `json:"html_url"`
	ID          int64  `json:"id"`
	Number      int64  `json:"number"`
	Title       string `json:"title"`
}

type GitHubPullRequest struct {
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
