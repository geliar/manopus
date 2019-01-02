package simple

import (
	"context"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/geliar/manopus/pkg/log"

	"github.com/geliar/manopus/pkg/payload"
	"github.com/geliar/manopus/pkg/processor"
)

func TestSimple(t *testing.T) {
	l := log.Output(ioutil.Discard)
	ctx := l.WithContext(context.Background())
	type args struct {
		ctx     context.Context
		config  *processor.ProcessorConfig
		payload *payload.Payload
	}
	tests := []struct {
		name       string
		p          *Simple
		args       args
		wantResult interface{}
		wantNext   processor.NextStatus
		wantErr    bool
	}{
		{
			name: "none",
			p:    &Simple{},
			args: args{
				ctx: ctx,
				config: &processor.ProcessorConfig{
					Encoding: "none",
					Script:   map[string]interface{}{"testkey": "test"},
				},
				payload: nil,
			},
			wantResult: map[string]interface{}{"testkey": "test"},
			wantNext:   processor.NextContinue,
			wantErr:    false,
		},
		{
			name: "json",
			p:    &Simple{},
			args: args{
				ctx: ctx,
				config: &processor.ProcessorConfig{
					Encoding: "json",
					Script:   map[string]interface{}{"testkey": "test"},
				},
				payload: nil,
			},
			wantResult: []byte("{\"testkey\":\"test\"}"),
			wantNext:   processor.NextContinue,
			wantErr:    false,
		},
		{
			name: "json error",
			p:    &Simple{},
			args: args{
				ctx: ctx,
				config: &processor.ProcessorConfig{
					Encoding: "json",
					Script:   map[string]interface{}{"testkey": make(chan int)},
				},
				payload: nil,
			},
			wantResult: nil,
			wantNext:   processor.NextStopSequence,
			wantErr:    true,
		},
		{
			name: "toml",
			p:    &Simple{},
			args: args{
				ctx: ctx,
				config: &processor.ProcessorConfig{
					Encoding: "toml",
					Script:   map[string]interface{}{"testkey": "test"},
				},
				payload: nil,
			},
			wantResult: []byte("testkey = \"test\"\n"),
			wantNext:   processor.NextContinue,
			wantErr:    false,
		},
		{
			name: "toml error",
			p:    &Simple{},
			args: args{
				ctx: ctx,
				config: &processor.ProcessorConfig{
					Encoding: "toml",
					Script:   map[string]interface{}{"testkey": make(chan int)},
				},
				payload: nil,
			},
			wantResult: nil,
			wantNext:   processor.NextStopSequence,
			wantErr:    true,
		},
		{
			name: "unknown encoder error",
			p:    &Simple{},
			args: args{
				ctx: ctx,
				config: &processor.ProcessorConfig{
					Encoding: "abc",
					Script:   map[string]interface{}{"testkey": make(chan int)},
				},
				payload: nil,
			},
			wantResult: nil,
			wantNext:   processor.NextStopSequence,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Simple{}
			gotResult, gotNext, err := p.Run(tt.args.ctx, tt.args.config, tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("Simple.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("Simple.Run() gotResult = %v, want %v", string(gotResult.([]byte)), string(tt.wantResult.([]byte)))
			}
			if !reflect.DeepEqual(gotNext, tt.wantNext) {
				t.Errorf("Simple.Run() gotNext = %v, want %v", gotNext, tt.wantNext)
			}
		})
	}
}
