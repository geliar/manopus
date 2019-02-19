package processor

import (
	"context"

	"github.com/geliar/manopus/pkg/payload"
)

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
	//Run execution of script
	Run(ctx context.Context, script interface{}, event *payload.Event, payload *payload.Payload) (next NextStatus, callback interface{}, responses []payload.Response, err error)
	//Match execution of match
	Match(ctx context.Context, match interface{}, payload *payload.Payload) (matched bool, err error)
}
