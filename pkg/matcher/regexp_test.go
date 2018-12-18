package matcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

type TestData struct {
	Matches []MatchConfig `yaml:"match"`
}

func TestRegExpMatcher(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		a := assert.New(t)
		data := []byte(`match:
  - field: some_field
    regexp: "(My|my) name is (?P<name>[A-Za-z]+)"
`)
		var result TestData
		a.NoError(yaml.Unmarshal(data, &result), "Cannot unmarshal test data", string(data))
		matches, matched := result.Matches[0].RegExp.Match("my name is John")
		a.True(matched)
		a.Equal("John", matches["name"])
	})
	t.Run("Not matched", func(t *testing.T) {
		a := assert.New(t)
		data := []byte(`match:
  - field: some_field
    regexp: "(My|my) name is (?P<name>[A-Za-z]+)"
`)
		var result TestData
		a.NoError(yaml.Unmarshal(data, &result), "Cannot unmarshal test data", string(data))
		matches, matched := result.Matches[0].RegExp.Match("my name John")
		a.False(matched)
		a.Len(matches, 0)
	})
	t.Run("Empty expression", func(t *testing.T) {
		a := assert.New(t)
		data := []byte(`match:
  - field: some_field
    regexp: ""
`)
		var result TestData
		a.EqualError(yaml.Unmarshal(data, &result), ErrRegExpEmpty.Error(), string(data))
	})
	t.Run("Wrong expression", func(t *testing.T) {
		a := assert.New(t)
		data := []byte(`match:
  - field: some_field
    regexp: "\l"
`)
		var result TestData
		a.Error(yaml.Unmarshal(data, &result), string(data))
	})
	t.Run("Expression is not a string", func(t *testing.T) {
		a := assert.New(t)
		data := []byte(`match:
  - field: some_field
    regexp:
      - some: field
`)
		var result TestData
		a.Error(yaml.Unmarshal(data, &result), string(data))
	})
}
