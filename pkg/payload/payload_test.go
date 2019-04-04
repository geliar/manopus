package payload

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/geliar/manopus/pkg/log"

	"github.com/stretchr/testify/assert"
)

func TestPayload_FromJson(t *testing.T) {
	l := log.Output(ioutil.Discard)
	ctx := l.WithContext(context.Background())
	tests := []struct {
		name string
		in   string
		out  Payload
	}{
		{"HappyPath", `{"req": {"testreq": "testreq is OK" },
                                 "env": {"testenv": "testenv is OK"},
                                 "export": {"testexport": "testexport is OK"},
                                 "match": {"testmatch": "testmatch is OK"}}`,
			Payload{
				Req:    map[string]interface{}{"testreq": "testreq is OK"},
				Env:    map[string]interface{}{"testenv": "testenv is OK"},
				Export: map[string]interface{}{"testexport": "testexport is OK"},
				Match:  map[string]interface{}{"testmatch": "testmatch is OK"},
			},
		},
		{"empty", "", Payload{}},
		{"emptyJSON", "{}", Payload{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			p := Payload{}
			p.FromJson(ctx, []byte(tt.in))
			a.EqualValues(tt.out, p)
		})
	}
}

func TestPayload_ToJson(t *testing.T) {
	l := log.Output(ioutil.Discard)
	ctx := l.WithContext(context.Background())
	tests := []struct {
		name string
		in   Payload
		out  []byte
	}{
		{"HappyPath", Payload{
			Event:  &EventInfo{Type: "some_type", Input: "some_input"},
			Req:    map[string]interface{}{"testreq": "testreq is OK"},
			Resp:   map[string]interface{}{"testresp": "testresp is OK"},
			Env:    map[string]interface{}{"testenv": "testenv is OK"},
			Export: map[string]interface{}{"testexport": "testexport is OK"},
			Match:  map[string]interface{}{"testmatch": "testmatch is OK"},
			Vars:   map[string]interface{}{"testvars": "testvars is OK"},
		},
			[]byte("{\"event\":{\"input\":\"some_input\",\"type\":\"some_type\"},\"env\":{\"testenv\":\"testenv is OK\"},\"vars\":{\"testvars\":\"testvars is OK\"},\"req\":{\"testreq\":\"testreq is OK\"},\"resp\":{\"testresp\":\"testresp is OK\"},\"export\":{\"testexport\":\"testexport is OK\"},\"match\":{\"testmatch\":\"testmatch is OK\"}}"),
		},
		{"emptyPayload", Payload{}, []byte("{\"event\":null,\"env\":null,\"vars\":null,\"req\":null,\"resp\":null,\"export\":null,\"match\":null}")},
		{"badPayload", Payload{Req: map[string]interface{}{"bad": make(chan int)}}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			p := tt.in
			buf := p.ToJson(ctx)
			a.EqualValues(string(tt.out), string(buf))
		})
	}
}

func TestPayload_QueryField(t *testing.T) {
	l := log.Output(ioutil.Discard)
	ctx := l.WithContext(context.Background())
	tests := []struct {
		name  string
		in    Payload
		query string
		out   interface{}
	}{
		{"HappyPath", Payload{
			Event:  &EventInfo{Type: "some_type", Input: "some_input"},
			Req:    map[string]interface{}{"testreq": "testreq is OK"},
			Env:    map[string]interface{}{"testenv": "testenv is OK"},
			Export: map[string]interface{}{"testexport": "testexport is OK"},
			Match:  map[string]interface{}{"testmatch": "testmatch is OK"},
		},
			"req.testreq", "testreq is OK",
		},
		{"emptyPayload", Payload{}, "req.testreq", nil},
		{"badPayload", Payload{Req: map[string]interface{}{"bad": make(chan int)}}, "req.testreq", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			p := tt.in
			val := p.QueryField(ctx, tt.query)
			a.EqualValues(tt.out, val)
		})
	}
}

func TestPayload_SetField(t *testing.T) {
	l := log.Output(ioutil.Discard)
	ctx := l.WithContext(context.Background())
	ch := make(chan int)
	tests := []struct {
		name  string
		in    Payload
		query string
		value interface{}
		out   Payload
	}{
		{"Existing data, same type", Payload{
			Req:    map[string]interface{}{"testreq": "testreq is OK"},
			Env:    map[string]interface{}{"testenv": "testenv is OK"},
			Export: map[string]interface{}{"testexport": "testexport is OK"},
			Match:  map[string]interface{}{"testmatch": "testmatch is OK"},
		},
			"req.testreq",
			"new data",
			Payload{
				Req:    map[string]interface{}{"testreq": "new data"},
				Env:    map[string]interface{}{"testenv": "testenv is OK"},
				Export: map[string]interface{}{"testexport": "testexport is OK"},
				Match:  map[string]interface{}{"testmatch": "testmatch is OK"},
			},
		},
		{"Existing data, new type bool", Payload{
			Req:    map[string]interface{}{"testreq": "testreq is OK"},
			Env:    map[string]interface{}{"testenv": "testenv is OK"},
			Export: map[string]interface{}{"testexport": "testexport is OK"},
			Match:  map[string]interface{}{"testmatch": "testmatch is OK"},
		},
			"req.testreq",
			true,
			Payload{
				Req:    map[string]interface{}{"testreq": true},
				Env:    map[string]interface{}{"testenv": "testenv is OK"},
				Export: map[string]interface{}{"testexport": "testexport is OK"},
				Match:  map[string]interface{}{"testmatch": "testmatch is OK"},
			},
		},
		{"Existing data, new type map", Payload{
			Req:    map[string]interface{}{"testreq": "testreq is OK"},
			Env:    map[string]interface{}{"testenv": "testenv is OK"},
			Export: map[string]interface{}{"testexport": "testexport is OK"},
			Match:  map[string]interface{}{"testmatch": "testmatch is OK"},
		},
			"req.testreq",
			map[string]interface{}{"test": "testvalue"},
			Payload{
				Req:    map[string]interface{}{"testreq": map[string]interface{}{"test": "testvalue"}},
				Env:    map[string]interface{}{"testenv": "testenv is OK"},
				Export: map[string]interface{}{"testexport": "testexport is OK"},
				Match:  map[string]interface{}{"testmatch": "testmatch is OK"},
			},
		},
		{"Existing data, new value is nil", Payload{
			Req:    map[string]interface{}{"testreq": "testreq is OK"},
			Env:    map[string]interface{}{"testenv": "testenv is OK"},
			Export: map[string]interface{}{"testexport": "testexport is OK"},
			Match:  map[string]interface{}{"testmatch": "testmatch is OK"},
		},
			"req.testreq",
			nil,
			Payload{
				Req:    map[string]interface{}{"testreq": nil},
				Env:    map[string]interface{}{"testenv": "testenv is OK"},
				Export: map[string]interface{}{"testexport": "testexport is OK"},
				Match:  map[string]interface{}{"testmatch": "testmatch is OK"},
			},
		},
		{"Bad existing data",
			Payload{
				Req: map[string]interface{}{"bad": ch},
			},
			"req.testreq",
			"test",
			Payload{
				Req: map[string]interface{}{"bad": ch},
			},
		},
		{"Bad new data",
			Payload{
				Req:    map[string]interface{}{"testreq": "testreq is OK"},
				Env:    map[string]interface{}{"testenv": "testenv is OK"},
				Export: map[string]interface{}{"testexport": "testexport is OK"},
				Match:  map[string]interface{}{"testmatch": "testmatch is OK"},
			},
			"req.testreq",
			ch,
			Payload{
				Req:    map[string]interface{}{"testreq": "testreq is OK"},
				Env:    map[string]interface{}{"testenv": "testenv is OK"},
				Export: map[string]interface{}{"testexport": "testexport is OK"},
				Match:  map[string]interface{}{"testmatch": "testmatch is OK"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			p := tt.in
			p.SetField(ctx, tt.query, tt.value)
			a.EqualValues(tt.out, p)
		})
	}
}

func TestPayload_ExportField(t *testing.T) {
	l := log.Output(ioutil.Discard)
	ctx := l.WithContext(context.Background())
	tests := []struct {
		name    string
		in      Payload
		current string
		new     string
		out     Payload
	}{
		{"Export from req", Payload{
			Req:    map[string]interface{}{"testreq": "oldreq"},
			Env:    map[string]interface{}{"testenv": "oldenv"},
			Export: map[string]interface{}{},
			Match:  map[string]interface{}{"testmatch": "oldmatch"},
		},
			"req.testreq",
			"newexport",
			Payload{
				Req:    map[string]interface{}{"testreq": "oldreq"},
				Env:    map[string]interface{}{"testenv": "oldenv"},
				Export: map[string]interface{}{"newexport": "oldreq"},
				Match:  map[string]interface{}{"testmatch": "oldmatch"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			p := tt.in
			p.ExportField(ctx, tt.current, tt.new)
			a.EqualValues(tt.out, p)
		})
	}
}
