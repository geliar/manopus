package payload

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

//Payload is the struct with all context sequences
type Payload struct {
	Event  *EventInfo             `yaml:"event" json:"event"`
	Env    map[string]interface{} `yaml:"env" json:"env"`
	Vars   map[string]interface{} `yaml:"vars" json:"vars"`
	Req    interface{}            `yaml:"req" json:"req"`
	Resp   map[string]interface{} `yaml:"resp" json:"resp"`
	Export map[string]interface{} `yaml:"export" json:"export"`
	Match  map[string]interface{} `yaml:"match" json:"match"`
}

//EventInfo describes event information data
type EventInfo struct {
	Input string `yaml:"input" json:"input"`
	Type  string `yaml:"type" json:"type"`
}

//ToJSON converts Payload to binary JSON data
func (p *Payload) ToJSON(ctx context.Context) []byte {
	l := logger(ctx)
	buf, err := json.Marshal(p)
	if err != nil {
		l.Error().Err(err).Msg("Cannot marshal payload to JSON")
		return nil
	}
	return buf
}

//FromJSON converts binary JSON data to Payload
func (p *Payload) FromJSON(ctx context.Context, data []byte) *Payload {
	l := logger(ctx)
	err := json.Unmarshal(data, p)
	if err != nil {
		l.Error().Err(err).Msg("Cannot parse JSON to payload")
		return nil
	}
	return p
}

//QueryField gets data from specified field in payload
func (p *Payload) QueryField(ctx context.Context, query string) interface{} {
	return gjson.GetBytes(p.ToJSON(ctx), query).Value()
}

//SetField sets specified field with data in payload
func (p *Payload) SetField(ctx context.Context, query string, data interface{}) {
	l := logger(ctx).With().Str("query", query).Logger()
	buf := p.ToJSON(ctx)
	if buf == nil {
		return
	}
	buf, err := sjson.SetBytes(buf, query, data)
	if err != nil {
		l.Error().Msg("Cannot set field value")
		return
	}
	p.FromJSON(ctx, buf)
}

//ExportField adds specified field to export map
func (p *Payload) ExportField(ctx context.Context, current string, new string) {
	value := p.QueryField(ctx, current)
	p.SetField(ctx, fmt.Sprintf("export.%s", new), value)
}
