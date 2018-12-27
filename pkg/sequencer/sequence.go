package sequencer

import (
	"context"

	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/payload"
)

type Sequence struct {
	sequenceConfig SequenceConfig
	step           int
	payload        *payload.Payload
}

func (s *Sequence) Match(ctx context.Context, defaultInputs []string, event *input.Event) bool {
	newpayload := *(s.payload)
	newpayload.Req = event.Data

	if len(s.sequenceConfig.Steps[s.step].Inputs) > 0 {
		if !contains(s.sequenceConfig.Steps[s.step].Inputs, event.Input) {
			return false
		}
	} else {
		if !contains(defaultInputs, event.Input) {
			return false
		}
	}

	for i := range s.sequenceConfig.Steps[s.step].Match {
		if !s.sequenceConfig.Steps[s.step].Match[i].Match(ctx, &newpayload) {
			return false
		}
	}

	*(s.payload) = newpayload
	return true
}

func contains(s []string, str string) bool {
	for i := range s {
		if s[i] == str {
			return true
		}
	}
	return false
}
