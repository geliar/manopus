package processor

import (
	"context"
	"errors"

	"github.com/geliar/manopus/pkg/payload"
)

var ErrParseScript = errors.New("cannot parse script")

type NextStatus int

const (
	NextContinue NextStatus = iota
	NextStopSequence
	NextRepeatStep
)

//Processor represents interface of script executor
type Processor interface {
	//Type get type of the Processor
	Type() string
	//Run execution of processor
	Run(ctx context.Context, config *ProcessorConfig, payload *payload.Payload) (result interface{}, next NextStatus, err error)
}

type ProcessorConfig struct {
	//Type of the processor to execute
	Type string `yaml:"type"`
	//MaxExecutionTime (optional) maximum time (in seconds) a step can execute for
	MaxExecutionTime int64 `yaml:"max_execution_time"`
	// Encoding of response (none, plain, json, json, toml)
	Encoding string `yaml:"encoding"`
	// Destination defines which output should be triggered with processor response
	Destination string `yaml:"destination"`
	//SkipCallback step should skip returning response to input requester
	SkipCallback bool `yaml:"skip_callback"`
	//Script will be executed by processor
	Script interface{} `yaml:"script"`
	//Export list of variables to be exported after execution of step
	Export []struct {
		//Current variable name in payload
		Current string `yaml:"current"`
		//New variable name in export part of payload
		New string `yaml:"new"`
	} `yaml:"export"`
}
