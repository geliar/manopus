package payload

// Event input event structure
type Event struct {
	Input string
	Type  string
	ID    string
	Data  interface{}
}

// Response output response structure
type Response struct {
	// ID of original request
	ID string
	// Output where to send response
	Output string
	// Data response data
	Data map[string]interface{}
	// Request original request
	Request *Event
}
