package payload

import (
	"context"
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

func (p *Payload) ToJson(ctx context.Context) []byte {
	l := logger(ctx)
	buf, err := json.Marshal(p)
	if err != nil {
		l.Error().Err(err).Msg("Cannot marshal payload to JSON")
		return nil
	}
	return buf
}

func (p *Payload) FromJson(ctx context.Context, data []byte) *Payload {
	l := logger(ctx)
	err := json.Unmarshal(data, p)
	if err != nil {
		l.Error().Err(err).Msg("Cannot parse JSON to payload")
		return nil
	}
	return p
}

func (p *Payload) QueryField(ctx context.Context, query string) interface{} {
	return gjson.GetBytes(p.ToJson(ctx), query).Value()
}

func (p *Payload) SetField(ctx context.Context, query string, data interface{}) {
	l := logger(ctx).With().Str("query", query).Logger()
	buf := p.ToJson(ctx)
	if buf == nil {
		return
	}
	buf, err := sjson.SetBytes(buf, query, data)
	if err != nil {
		l.Error().Msg("Cannot set field value")
		return
	}
	p.FromJson(ctx, buf)
}
