package matcher

type MatchConfig struct {
	Field string `yaml:"field"`
	CompareField string `yaml:"compare_field"`
	Operator string `yaml:"operator"`
	RegExp *RegExpMatcher `yaml:"regexp"`
	Value interface{} `yaml:"value"`
}

