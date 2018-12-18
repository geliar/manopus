package payload

import (
	"encoding/json"

	"github.com/tidwall/sjson"

	"github.com/tidwall/gjson"
)

type Payload struct {
	Env    map[string]interface{} `yaml:"env" json:"env"`
	Req    map[string]interface{} `yaml:"req" json:"req"`
	Export map[string]interface{} `yaml:"export" json:"export"`
	Match  map[string]interface{} `yaml:"match" json:"match"`
}

func (p *Payload) ToJson() []byte {
	l := logger()
	buf, err := json.Marshal(p)
	if err != nil {
		l.Error().Err(err).Msg("Cannot marshal payload to JSON")
		return nil
	}
	return buf
}

func (p *Payload) FromJson(data []byte) *Payload {
	l := logger()
	err := json.Unmarshal(data, p)
	if err != nil {
		l.Error().Err(err).Msg("Cannot parse JSON to payload")
		return nil
	}
	return p
}

func (p *Payload) QueryField(query string) interface{} {
	return gjson.GetBytes(p.ToJson(), query).Value()
}

func (p *Payload) SetField(query string, data interface{}) {
	l := logger().With().Str("query", query).Logger()
	buf := p.ToJson()
	if buf == nil {
		return
	}
	buf, err := sjson.SetBytes(buf, query, data)
	if err != nil {
		l.Error().Msg("Cannot set field value")
		return
	}
	p.FromJson(buf)
}
