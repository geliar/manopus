package matcher

import (
	"context"
	"reflect"

	"github.com/geliar/manopus/pkg/payload"
)

// MatchConfig contains structure of the 'match' list
type MatchConfig struct {
	//And is used when you need to check whether all of the submatches are true
	And []MatchConfig `yaml:"and"`
	//Or is used when you need to check whether at least one of the submatches is true
	Or []MatchConfig `yaml:"or"`
	//Xor is used when you need to check whether only one of the submatches is true
	Xor []MatchConfig `yaml:"xor"`
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
	//Negative is true when result of the match should be reverted !=
	Negative bool `yaml:"negative"`
}

// Match matches payload with internal matching rules
func (m *MatchConfig) Match(ctx context.Context, payload *payload.Payload) bool {
	and := true
	if len(m.And) > 0 {
		and = m.And[0].Match(ctx, payload)
		if len(m.And) > 1 {
			for i := 1; i < len(m.And); i++ {
				and = and && m.And[i].Match(ctx, payload)
			}
		}
	}
	or := true
	if len(m.Or) > 0 {
		or = m.Or[0].Match(ctx, payload)
		if len(m.Or) > 1 {
			for i := 1; i < len(m.Or); i++ {
				or = or || m.Or[i].Match(ctx, payload)
			}
		}
	}
	xor := true
	if len(m.Xor) > 0 {
		xor = m.Xor[0].Match(ctx, payload)
		if len(m.Xor) > 1 {
			for i := 1; i < len(m.Xor); i++ {
				xor = xor != m.Xor[i].Match(ctx, payload)
			}
		}
	}
	field := true
	if m.Field != "" {
		field = m.matchField(ctx, payload)
	}
	result := and && or && xor && field
	return result != m.Negative
}

func (m *MatchConfig) matchField(ctx context.Context, payload *payload.Payload) bool {
	l := logger(ctx)

	f := payload.QueryField(ctx, m.Field)
	if f == nil {
		l.Debug().
			Str("match_field", m.Field).
			Msg("Cannot find such field in payload")
		return false
	}
	// If field equals to 'value' field
	if m.Value != nil {
		return m.compare(ctx, f, m.Value, m.Operator)
	}
	if m.Compare != "" {
		c := payload.QueryField(ctx, m.Compare)
		if c == nil {
			l.Debug().
				Str("match_compare", m.Compare).
				Msg("Cannot find such compare field in payload")
			return false
		}
		return m.compare(ctx, f, c, m.Operator)
	}
	switch v := f.(type) {
	case string:
		// If match has regexp expression
		if m.RegExp != nil {
			matches, matched := m.RegExp.Match(v)
			if !matched {
				return false
			}
			// Merging matches
			for k := range matches {
				if payload.Match == nil {
					payload.Match = map[string]interface{}{}
				}
				payload.Match[k] = matches[k]
			}
			return true
		}
	}
	return false
}

func (m *MatchConfig) compare(ctx context.Context, f, c interface{}, operator string) bool {
	if operator == "" {
		return reflect.DeepEqual(f, c)
	}
	var a, b float64

	// f should be always float64 because of gjson package
	switch v := f.(type) {
	case float64:
		a = v
	default:
		return false
	}
	// Converting c to float64
	switch v := c.(type) {
	case int:
		b = float64(v)
	case float64:
		b = v
	default:
		return false
	}

	//Comparing
	switch operator {
	case "==", "=", "eq":
		return a == b
	case "!=", "ne":
		return a != b
	case ">", "gt":
		return a > b
	case ">=", "ge":
		return a >= b
	case "<", "lt":
		return a < b
	case "<=", "le":
		return a <= b
	}
	return false
}
