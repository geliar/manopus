package contextdata

type ContextData struct {
	Env    map[string]interface{} `yaml:"env" json:"env"`
	Req    map[string]interface{} `yaml:"req" json:"req"`
	Export map[string]interface{} `yaml:"export" json:"export"`
}
