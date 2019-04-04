package slack

const RequestTypeInteraction = "interaction"

type RequestInteraction struct {
	SlackMessage
	CallbackID string          `starlark:"callback_id" json:"callback_id"`
	Actions    []MessageAction `starlark:"actions" json:"actions"`
}

const RequestTypeMessage = "event"

type RequestMessage struct {
	SlackMessage
	Mentioned bool `starlark:"mentioned" json:"mentioned"`
}

type SlackMessage struct {
	UserID          string `starlark:"user_id" json:"user_id"`
	UserName        string `starlark:"user_name" json:"user_name"`
	UserDisplayName string `starlark:"user_display_name" json:"user_display_name"`
	UserRealName    string `starlark:"user_real_name" json:"user_real_name"`
	ChannelID       string `starlark:"channel_id" json:"channel_id"`
	ChannelName     string `starlark:"channel_name" json:"channel_name"`
	ThreadTS        string `starlark:"thread_ts" json:"thread_ts"`
	Timestamp       string `starlark:"timestamp" json:"timestamp"`
	Message         string `starlark:"message" json:"message"`
	Direct          bool   `starlark:"direct" json:"direct"`
}

type MessageAction struct {
	Name  string `starlark:"name" json:"name"`
	Value string `starlark:"value" json:"value"`
	Text  string `starlark:"text" json:"text"`
}
