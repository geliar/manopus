package report

// Config report configuration structure
type Config struct {
	Driver string                 `json:"driver"`
	Config map[string]interface{} `json:"config"`
}
