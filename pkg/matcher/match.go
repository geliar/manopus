package matcher

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
