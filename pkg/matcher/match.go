package matcher

import "github.com/geliar/manopus/pkg/payload"

type MatchConfig struct {
	//Field field to basic compare
	Field string `yaml:"field"`
	//Compare field to compare with
	Compare string `yaml:"compare"`
	//Value value to compare with
	Value interface{} `yaml:"value"`
	//Operator comparison operator (e.g. "eq", "==", "lt", "<", "gt", ">" and)
	Operator string `yaml:"operator"`
	//RegExp regexp to match with and collect values to match object
	RegExp *RegExpMatcher `yaml:"regexp"`
}

func (m *MatchConfig) Match(payload *payload.Payload) (matched bool) {
	return false
}
