package github

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/davecgh/go-spew/spew"

	cgithub "github.com/google/go-github/v24/github"
	whgithub "gopkg.in/go-playground/webhooks.v5/github"

	"github.com/rs/zerolog/hlog"

	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/payload"
)

type GitHub struct {
	created  int64
	id       int64
	name     string
	handlers []input.Handler
	stop     bool
	stopCh   chan struct{}
	mu       sync.RWMutex
	hook     *whgithub.Webhook
	client   *cgithub.Client
}

func (c *GitHub) Name() string {
	return c.name
}

func (c *GitHub) Type() string {
	return serviceName
}

func (c *GitHub) RegisterHandler(ctx context.Context, handler input.Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers = append(c.handlers, handler)
}

func (c *GitHub) Send(ctx context.Context, response *payload.Response) map[string]interface{} {
	l := logger(ctx)
	l.Debug().
		Str("input_name", response.Request.Input).
		Str("input_event_id", response.Request.ID).
		Msg("Received Send event")
	if response.Data == nil {
		l.Error().Msg("Data field of request is empty")
		return nil
	}
	f, _ := response.Data["function"].(string)

	switch f {
	case "":
		l.Error().Msg("function field is empty")
		return nil
	case "pull_request_merge":
		owner, _ := response.Data["repo_owner"].(string)
		repo, _ := response.Data["repo_name"].(string)
		number, _ := response.Data["pr_number"].(int64)
		message, _ := response.Data["merge_message"].(string)
		title, _ := response.Data["merge_title"].(string)
		method, _ := response.Data["merge_method"].(string)
		l = l.With().
			Str("repo_name", repo).
			Int64("pr_number", number).
			Str("repo_owner", owner).
			Logger()
		if owner == "" || repo == "" || number == 0 {
			l.Error().Msg("Required fields are empty")
			return map[string]interface{}{
				"result": false,
			}
		}
		if method == "" {
			method = "merge"
		}
		pr, _, err := c.client.PullRequests.Get(ctx, owner, repo, int(number))
		if err != nil {
			l.Error().Err(err).Msg("Error calling PullRequests.Get on GitHub API")
			return map[string]interface{}{
				"result": false,
			}
		}
		if pr == nil {
			l.Error().Msg("Cannot get Pull Request data")
			return map[string]interface{}{
				"result": false,
			}
		}
		if pr.Mergeable == nil || !*pr.Mergeable ||
			pr.MergeCommitSHA == nil || *(pr.MergeCommitSHA) == "" {
			l.Error().Msg("Pull Request is not mergeable or not found")
			return map[string]interface{}{
				"result": false,
			}
		}
		opts := &cgithub.PullRequestOptions{
			SHA:         *(pr.Head.SHA),
			CommitTitle: title,
			MergeMethod: method,
		}
		spew.Dump(opts)
		result, _, err := c.client.PullRequests.Merge(ctx, owner, repo, int(number), message, opts)
		if err != nil && strings.Contains(err.Error(), "409 Head branch was modified. Review and try the merge again") {
			result, _, err = c.client.PullRequests.Merge(ctx, owner, repo, int(number), message, opts)
		}
		if err != nil {
			l.Error().Err(err).Msg("Error calling PullRequests.Merge on GitHub API")
			return map[string]interface{}{
				"result": false,
			}
		}
		return map[string]interface{}{
			"merged":        result.GetMerged(),
			"merge_message": result.GetMessage(),
			"sha":           result.SHA,
			"result":        true,
		}
	case "issue_comment":
		owner, _ := response.Data["repo_owner"].(string)
		repo, _ := response.Data["repo_name"].(string)
		number, _ := response.Data["issue_number"].(int64)
		message := response.Data["message"].(string)

		l = l.With().
			Str("repo_owner", owner).
			Str("repo_name", repo).
			Int64("issue_number", number).
			Str("tag_message", message).
			Logger()
		if owner == "" || repo == "" || number == 0 || message == "" {
			l.Error().Msg("Required fields are empty")
			return map[string]interface{}{
				"result": false,
			}
		}
		comment := &cgithub.IssueComment{
			Body: &message,
		}
		result, _, err := c.client.Issues.CreateComment(ctx, owner, repo, int(number), comment)
		if err != nil {
			l.Error().Err(err).Msg("Error calling Issues.CreateComment on GitHub API")
			return map[string]interface{}{
				"result": false,
			}
		}

		return map[string]interface{}{
			"result":     true,
			"html_url":   result.GetHTMLURL(),
			"body":       result.GetBody(),
			"sender":     result.GetUser().Login,
			"sender_url": result.GetUser().HTMLURL,
		}
	case "create_tag":
		owner, _ := response.Data["repo_owner"].(string)
		repo, _ := response.Data["repo_name"].(string)
		object, _ := response.Data["tag_object"].(string)
		tagName, _ := response.Data["tag_name"].(string)
		tagType, _ := response.Data["tag_type"].(string)
		message, _ := response.Data["tag_message"].(string)
		taggerName, _ := response.Data["tagger_name"].(string)
		taggerEmail, _ := response.Data["tagger_email"].(string)
		/*taggerDate := response.Data["tag_date"].(string) //Optional
		if taggerDate == "" {
			taggerDate = time.Now().UTC().Format("2006-01-02T15:04:05-0700")
		}*/
		if tagType == "" {
			tagType = "commit"
		}
		l = l.With().
			Str("repo_owner", owner).
			Str("repo_name", repo).
			Str("tag_object", object).
			Str("tag_name", tagName).
			Str("tag_type", tagType).
			Str("tag_message", message).
			Str("tagger_name", taggerName).
			Str("tagger_email", taggerEmail).
			//Str("tagger_date", taggerDate).
			Logger()
		if owner == "" || repo == "" || object == "" ||
			tagName == "" || message == "" || taggerName == "" || taggerEmail == "" {
			l.Error().Msg("Required fields are empty")
			return map[string]interface{}{
				"result": false,
			}
		}
		obj := cgithub.GitObject{
			SHA:  &object,
			Type: &tagType,
		}
		now := time.Now()
		author := cgithub.CommitAuthor{
			Date:  &now,
			Name:  &taggerName,
			Email: &taggerEmail,
		}
		tag := &cgithub.Tag{
			Tag:     &tagName,
			Object:  &obj,
			Tagger:  &author,
			Message: &message,
		}
		l.Debug().Msg("Creating tag")
		resTag, _, err := c.client.Git.CreateTag(ctx, owner, repo, tag)
		if err != nil {
			l.Error().Err(err).Msg("Error calling Git.CreateTag on GitHub API")
			return map[string]interface{}{
				"result": false,
			}
		}
		spew.Dump(resTag)
		refPath := "refs/tags/" + tagName
		ref := cgithub.Reference{
			Ref:    &refPath,
			Object: &cgithub.GitObject{SHA: resTag.SHA},
		}
		resRef, _, err := c.client.Git.CreateRef(ctx, owner, repo, &ref)
		if err != nil {
			l.Error().Err(err).Msg("Error calling Git.CreateRef on GitHub API")
			return map[string]interface{}{
				"result": false,
			}
		}
		return map[string]interface{}{
			"ref":        resRef.GetRef(),
			"ref_url":    resRef.GetURL(),
			"tag_type":   resRef.GetObject().GetType(),
			"tag_object": resRef.GetObject().GetSHA(),
			"tag_url":    resRef.GetObject().GetURL(),
			"result":     true,
		}
	}
	return nil
}

func (c *GitHub) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	tl := hlog.FromRequest(r)
	ctx := tl.WithContext(context.Background())
	_ = ctx
	l := logger(ctx)
	data, err := c.hook.Parse(r, whgithub.ReleaseEvent, whgithub.PullRequestEvent, whgithub.PushEvent, whgithub.IssueCommentEvent, whgithub.PingEvent)
	if err != nil {
		l.Error().Err(err).Msg("Error parsing event")
		if err == whgithub.ErrEventNotFound {

		}
	}
	//spew.Dump(data)
	event := new(payload.Event)
	event.ID = c.getID()
	event.Input = c.name
	switch v := data.(type) {

	case whgithub.PullRequestPayload:

		var issue GitHubIssue
		if parts := strings.Split(v.PullRequest.IssueURL, "/"); len(parts) == 8 &&
			parts[6] == "issues" &&
			parts[7] != "" {
			number, err := strconv.Atoi(parts[7])
			if err == nil {
				issue.URL = v.PullRequest.IssueURL
				issue.Number = int64(number)
			}
		}
		event.Type = RequestTypePullRequest
		event.Data = RequestPullRequest{
			PullRequestPayload: v,
			Issue:              issue,
		}

		l.Debug().Msg("Pull request event")
	case whgithub.IssueCommentPayload:
		var pr GitHubPullRequest
		if parts := strings.Split(v.Issue.HTMLURL, "/"); len(parts) == 7 &&
			parts[5] == "pull" &&
			parts[6] != "" {
			number, err := strconv.Atoi(parts[6])
			if err == nil {
				pr.HTMLURL = v.Issue.HTMLURL
				pr.Number = int64(number)
			}
		}
		event.Type = RequestTypeIssueComment
		event.Data = RequestIssueComment{
			IssueCommentPayload: v,
			GitHubPullRequest:   pr,
		}

		l.Debug().Msg("Issue comment event")
	case whgithub.PushPayload:
		var branch string
		if parts := strings.Split(v.Ref, "/")[2]; len(parts) == 3 {
			branch = strings.Split(v.Ref, "/")[2]
		}
		event.Type = RequestTypePush
		event.Data = RequestPush{
			PushPayload: v,
			Branch:      branch,
			Head: GitHubHead{
				Ref:  v.Ref,
				Sha:  v.HeadCommit.ID,
				User: v.HeadCommit.Author,
			},
		}

		l.Debug().Msg("Push event")

	case whgithub.PingPayload:
		l.Debug().Msg("Ping event")
	}
	if event.Data != nil {
		c.sendEventToHandlers(ctx, event)
	}
}

func (c *GitHub) Stop(ctx context.Context) {
	if !c.stop {
		c.stop = true
		close(c.stopCh)
	}
}

func (c *GitHub) getID() string {
	id := atomic.AddInt64(&c.id, 1)
	return fmt.Sprintf("%s-%d-%d", c.name, c.created, id)
}

func (c *GitHub) sendEventToHandlers(ctx context.Context, event *payload.Event) (response *payload.Response) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.handlers {
		resp := h(ctx, event)
		if resp != nil {
			response = new(payload.Response)
			response.Data = map[string]interface{}{
				"data": resp,
			}
			response.ID = event.ID
			response.Request = event
			response.Output = serviceName
		}
	}
	return
}
