package matcher

import (
	"context"
	"reflect"

	"github.com/davecgh/go-spew/spew"

	"github.com/geliar/manopus/pkg/payload"
)

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

func (m *MatchConfig) Match(ctx context.Context, payload *payload.Payload) bool {
	l := logger(ctx)
	if m.Field == "" {
		l.Error().
			Msg("'field' in 'match' should not be empty")
		return false
	}
	f := payload.QueryField(ctx, m.Field)
	spew.Dump(f, m.Value)
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
				Str("match_compare", m.Field).
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

	// Converting f to float64
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
