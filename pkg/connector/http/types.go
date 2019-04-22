package http

const requestTypeHTTPRequest = "request"

type requestHTTPRequest struct {
	Method      string              `json:"method" starlark:"method"`
	Host        string              `json:"host" starlark:"host"`
	RemoteAddr  string              `json:"remote_addr" starlark:"remote_addr"`
	URI         string              `json:"uri" starlark:"uri"`
	Path        string              `json:"path" starlark:"path"`
	Form        map[string][]string `json:"form" starlark:"form"`
	ContentType string              `json:"content_type" starlark:"content_type"`
	Referer     string              `json:"referer" starlark:"referer"`
	UserAgent   string              `json:"user_agent" starlark:"user_agent"`
	Headers     map[string][]string `json:"headers" starlark:"headers"`
	Body        string              `json:"body" starlark:"body"`
}

const requestTypeHTTPJSONRequest = "json_request"

type requestHTTPJSONRequest struct {
	requestHTTPRequest
	JSON map[string]interface{} `json:"json" starlark:"json"`
}
