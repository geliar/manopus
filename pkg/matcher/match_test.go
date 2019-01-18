package matcher

import (
	"context"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/geliar/manopus/pkg/payload"
	"github.com/geliar/yaml"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func init() {
	yaml.DefaultMapType = reflect.TypeOf(map[string]interface{}{})
}

type TestData struct {
	Matches []MatchConfig `yaml:"match"`
}

func TestMatchConfig_Match(t *testing.T) {
	l := log.Output(ioutil.Discard)
	ctx := l.WithContext(context.Background())
	tests := []struct {
		name    string
		in      payload.Payload
		match   string
		out     payload.Payload
		matched bool
	}{
		{"ValueStringMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testdata"},
			},
			`
match:
  - field: req.testreq
    value: "testdata"
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testdata"},
			},
			true,
		},
		{"ValueStringNotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testdata"},
			},
			`
match:
  - field: req.testreq
    value: "wrongdata"
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testdata"},
			},
			false,
		},
		{"ValueStringMatchedNegative",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testdata"},
			},
			`
match:
  - field: req.testreq
    value: "testdata"
    negative: true
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testdata"},
			},
			false,
		},
		{"ValueBoolMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": true},
			},
			`
match:
  - field: req.testreq
    value: true
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": true},
			},
			true,
		},
		{"ValueBoolNotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": true},
			},
			`
match:
  - field: req.testreq
    value: false
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": true},
			},
			false,
		},
		{"ValueInt<Matched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": 10},
			},
			`
match:
  - field: req.testreq
    value: 20
    operator: "<"
  - field: req.testreq
    value: 11
    operator: "lt"
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": 10},
			},
			true,
		},
		{"ValueInt<NotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": 30},
			},
			`
match:
  - field: req.testreq
    value: 20
    operator: "<"
  - field: req.testreq
    value: 11
    operator: "lt"
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": 30},
			},
			false,
		},
		{"ValueInt<=Matched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": 10},
			},
			`
match:
  - field: req.testreq
    value: 10
    operator: "<="
  - field: req.testreq
    value: 11
    operator: "le"
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": 10},
			},
			true,
		},
		{"ValueInt>Matched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": 30},
			},
			`
match:
  - field: req.testreq
    value: 10
    operator: ">"
  - field: req.testreq
    value: 29
    operator: "gt"
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": 30},
			},
			true,
		},
		{"ValueInt>=Matched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": 30},
			},
			`
match:
  - field: req.testreq
    value: 10
    operator: ">="
  - field: req.testreq
    value: 30
    operator: "ge"
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": 30},
			},
			true,
		},
		{"ValueInt==Matched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": 30},
			},
			`
match:
  - field: req.testreq
    value: 30
    operator: "=="
  - field: req.testreq
    value: 30
    operator: "eq"
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": 30},
			},
			true,
		},
		{"ValueInt!=Matched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": 0},
			},
			`
match:
  - field: req.testreq
    value: 30
    operator: "!="
  - field: req.testreq
    value: 30
    operator: "ne"
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": 0},
			},
			true,
		},
		{"ValueIntUnknownOperatorNotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": 30},
			},
			`
match:
  - field: req.testreq
    value: 30
    operator: "!!"
  - field: req.testreq
    value: 30
    operator: "<>"
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": 30},
			},
			false,
		},
		{"ValueFloat==Matched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": 30.1},
			},
			`
match:
  - field: req.testreq
    value: 30.1
    operator: "=="
  - field: req.testreq
    value: 30.1
    operator: "eq"
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": 30.1},
			},
			true,
		},
		{"ValueString==NotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": 30.1},
			},
			`
match:
  - field: req.testreq
    value: "s"
    operator: "=="
  - field: req.testreq
    value: s
    operator: "eq"
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": 30.1},
			},
			false,
		},
		{"FieldString==NotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "s"},
			},
			`
match:
  - field: req.testreq
    value: "s"
    operator: "=="
  - field: req.testreq
    value: s
    operator: "eq"
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "s"},
			},
			false,
		},
		{"ValueObjectMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": map[string]interface{}{"var": true}},
			},
			`
match:
  - field: req.testreq
    value:
      "var": true
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": map[string]interface{}{"var": true}},
			},
			true,
		},
		{"ValueObjectNotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": map[string]interface{}{"var": true}},
			},
			`
match:
  - field: req.testreq
    value:
      "var": false
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": map[string]interface{}{"var": true}},
			},
			false,
		},
		{"CompareMatched",
			payload.Payload{
				Req:    map[string]interface{}{"testreq": "testdata"},
				Export: map[string]interface{}{"testexport": "testdata"},
			},
			`
match:
  - field: req.testreq
    compare: export.testexport
`,
			payload.Payload{
				Req:    map[string]interface{}{"testreq": "testdata"},
				Export: map[string]interface{}{"testexport": "testdata"},
			},
			true,
		},
		{"RegexpMatchedNoOutputValues",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "My name is John"},
			},
			`
match:
  - field: req.testreq
    regexp: 'My name is [A-Za-z]+'
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "My name is John"},
			},
			true,
		},
		{"RegexpMatchedOutputValues",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "My name is John"},
			},
			`
match:
  - field: req.testreq
    regexp: '(My|my) name is (?P<name>[A-Za-z]+)'
`,
			payload.Payload{
				Req:   map[string]interface{}{"testreq": "My name is John"},
				Match: map[string]interface{}{"name": "John"},
			},
			true,
		},
		{"RegexpNotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "My name is !John"},
			},
			`
match:
  - field: req.testreq
    regexp: '(My|my) name is (?P<name>[A-Za-z]+)'
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "My name is !John"},
			},
			false,
		},
		{"RegexpValueIntNotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": 0},
			},
			`
match:
  - field: req.testreq
    regexp: '(My|my) name is (?P<name>[A-Za-z]+)'
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": 0},
			},
			false,
		},
		{"FieldIsEmpty",
			payload.Payload{},
			`
match:
  - field:
    value:
      "var": true
`,
			payload.Payload{},
			true,
		},
		{"NoSuchField",
			payload.Payload{},
			`
match:
  - field: req.abc
    value:
      "var": true
`,
			payload.Payload{},
			false,
		},
		{"NoSuchCompare",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testdata"},
			},
			`
match:
  - field: req.testreq
    compare: export.abc
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testdata"},
			},
			false,
		},
		{"ValueBoolAndOrXorMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": true},
			},
			`
match:
  - and:
    - field: req.testreq
      value: true
    - field: req.testreq
      value: true
    - field: req.testreq
      value: true
    or:
    - field: req.testreq
      value: false
    - field: req.testreq
      value: true
    - field: req.testreq
      value: true
    xor:
    - field: req.testreq
      value: true
    - field: req.testreq
      value: false
    - field: req.testreq
      value: false
    field: req.testreq
    value: true
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": true},
			},
			true,
		},
		{"Nothing to do",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testdata"},
			},
			`
match:
  - field: req.testreq
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testdata"},
			},
			false,
		},
		{"OperatorStringEmptyMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": ""},
			},
			`
match:
  - field: req.testreq
    operator: empty
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": ""},
			},
			true,
		},
		{"OperatorStringEmptyNotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "qwe"},
			},
			`
match:
  - field: req.testreq
    operator: empty
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "qwe"},
			},
			false,
		},
		{"OperatorStringNotEmptyMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "qwe"},
			},
			`
match:
  - field: req.testreq
    operator: not_empty
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "qwe"},
			},
			true,
		},
		{"OperatorStringNotEmptyNotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": ""},
			},
			`
match:
  - field: req.testreq
    operator: not_empty
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": ""},
			},
			false,
		},
		{"OperatorNotStringEmpty",
			payload.Payload{
				Req: map[string]interface{}{"testreq": 123},
			},
			`
match:
  - field: req.testreq
    operator: empty
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": 123},
			},
			false,
		},
		//Contains
		{"OperatorStringContainsMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			`
match:
  - field: req.testreq
    operator: contains
    value: match
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			true,
		},
		{"OperatorStringContainsNotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			`
match:
  - field: req.testreq
    operator: contains
    value: notmatch
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			false,
		},
		{"OperatorStringContainsEmptyNotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			`
match:
  - field: req.testreq
    operator: contains
    value: ""
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			false,
		},
		//HasPrefix
		{"OperatorStringHasPrefixMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			`
match:
  - field: req.testreq
    operator: has_prefix
    value: test
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			true,
		},
		{"OperatorStringHasPrefixNotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			`
match:
  - field: req.testreq
    operator: has_prefix
    value: match
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			false,
		},
		{"OperatorStringHasPrefixEmptyNotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			`
match:
  - field: req.testreq
    operator: has_prefix
    value: ""
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			false,
		},
		//HasSuffix
		{"OperatorStringHasSuffixMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			`
match:
  - field: req.testreq
    operator: has_suffix
    value: case
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			true,
		},
		{"OperatorStringHasSuffixNotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			`
match:
  - field: req.testreq
    operator: has_suffix
    value: match
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			false,
		},
		{"OperatorStringHasSuffixEmptyNotMatched",
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			`
match:
  - field: req.testreq
    operator: has_suffix
    value: ""
`,
			payload.Payload{
				Req: map[string]interface{}{"testreq": "testmatchcase"},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			p := tt.in
			var m TestData
			a.NoError(yaml.Unmarshal([]byte(tt.match), &m))
			for tm := range m.Matches {
				a.Equal(tt.matched, m.Matches[tm].Match(ctx, &p))
			}
			a.EqualValues(tt.out, p)
		})
	}
}
