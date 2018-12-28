package bash

import (
	"context"
	"io/ioutil"
	"sort"
	"strings"
	"testing"

	"github.com/geliar/manopus/pkg/log"

	"github.com/stretchr/testify/assert"
)

func Test_preparePayload(t *testing.T) {
	type args struct {
		prefix  string
		payload map[string]interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
	}{
		{
			"SimpleTypes",
			args{
				"ENV_",
				map[string]interface{}{
					"test-string": "some \" test",
					"test int":    123,
					"test  float": 0.123,
				},
			},
			`ENV_TEST_STRING="some \" test"
ENV_TEST_INT="123"
ENV_TEST__FLOAT="0.123"`,
		},
		{
			"Array",
			args{
				"ENV_",
				map[string]interface{}{
					"test-string": "some \" test",
					"test array": []interface{}{
						1,
						0.234,
						"abc",
						[]int{1, 2, 3},
					},
					"test  float": 0.123,
				},
			},
			`ENV_TEST_STRING="some \" test"
ENV_TEST_ARRAY[0]="1"
ENV_TEST_ARRAY[1]="0.234"
ENV_TEST_ARRAY[2]="abc"
ENV_TEST__FLOAT="0.123"`,
		},
		{
			"Map",
			args{
				"ENV_",
				map[string]interface{}{
					"test-string": "some \" test",
					"test map": map[string]interface{}{
						"test-string": "some additional test",
						"test array": []interface{}{
							2,
							0.987,
							"zxc",
							[]int{7, 8, 5},
						},
						"test  float": 0.1764,
					},
					"test  float": 0.123,
				},
			},
			`ENV_TEST_STRING="some \" test"
ENV_TEST_MAP_TEST_STRING="some additional test"
ENV_TEST_MAP_TEST_ARRAY[0]="2"
ENV_TEST_MAP_TEST_ARRAY[1]="0.987"
ENV_TEST_MAP_TEST_ARRAY[2]="zxc"
ENV_TEST_MAP_TEST__FLOAT="0.1764"
ENV_TEST__FLOAT="0.123"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			want := strings.Split(tt.wantResult+"\n", "\n")
			sort.Strings(want)
			got := strings.Split(preparePayload(tt.args.prefix, tt.args.payload), "\n")
			sort.Strings(got)
			a.Equal(want, got)
		})
	}
}

func TestBash_collectScript(t *testing.T) {
	l := log.Output(ioutil.Discard)
	ctx := l.WithContext(context.Background())
	type args struct {
		ctx    context.Context
		script interface{}
	}
	tests := []struct {
		name       string
		b          Bash
		args       args
		wantResult string
	}{
		{
			"String",
			Bash{},
			args{
				ctx,
				"some script",
			},
			"some script\n",
		},
		{
			"Array",
			Bash{},
			args{
				ctx,
				[]interface{}{
					"Some",
					"script",
					"is",
					"here",
				},
			},
			"Some\nscript\nis\nhere\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			b := Bash{}
			a.Equal(tt.wantResult, b.collectScript(tt.args.ctx, tt.args.script))
		})
	}
}
