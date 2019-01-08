package payload

type Event struct {
	Input string
	Type  string
	ID    string
	Data  map[string]interface{}
}

type Response struct {
	// ID of original request
	ID string
	// Type of response. Default value depends on connector defaults.
	Type string
	// Data response data
	Data map[string]interface{}
	// Encoding of response (none, plain, json, json, toml)
	Encoding string
	// Request original request
	Request *Event
}
