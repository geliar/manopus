package processor

import (
	"context"

	"github.com/geliar/manopus/pkg/payload"
	"github.com/geliar/manopus/pkg/report"
)

// NextStatus type of sequence statuses
type NextStatus int

const (
	// NextContinue sequencer should continue execution with the next step
	NextContinue NextStatus = iota
	// NextStopSequence sequencer should stop execution after this step
	NextStopSequence
	// NextRepeatStep sequencer should repeat this step one more time
	NextRepeatStep
)

//Processor represents interface of script executor
type Processor interface {
	//Type get type of the Processor
	Type() string
	//Run execution of script
	Run(ctx context.Context, reporter report.Driver, script interface{}, event *payload.Event, payload *payload.Payload) (next NextStatus, callback interface{}, responses []payload.Response, err error)
	//Match execution of match
	Match(ctx context.Context, match interface{}, payload *payload.Payload) (matched bool, err error)
}
