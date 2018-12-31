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
	//Script will be executed by processor
	Script interface{} `yaml:"script"`
}
