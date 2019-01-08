package sequencer

import (
	"context"
	"time"

	"github.com/geliar/manopus/pkg/payload"
)

type Sequence struct {
	sequenceConfig SequenceConfig
	step           int
	payload        *payload.Payload
	latestMatch    time.Time
}

func (s *Sequence) Match(ctx context.Context, defaultInputs []string, event *payload.Event) bool {
	l := logger(ctx)
	l = l.With().
		Str("sequence_name", s.sequenceConfig.Name).
		Int("sequence_step", s.step).Logger()
	l.Debug().Msg("Matching")
	newpayload := *(s.payload)
	newpayload.Req = event.Data
	ctx = l.WithContext(ctx)
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
	s.latestMatch = time.Now()
	return true
}

func (s *Sequence) TimedOut(ctx context.Context) bool {
	l := logger(ctx)
	l = l.With().
		Str("sequence_name", s.sequenceConfig.Name).
		Int("sequence_step", s.step).Logger()
	if s.sequenceConfig.Steps[s.step].Timeout > 0 && time.Now().After(s.latestMatch.Add(time.Duration(s.sequenceConfig.Steps[s.step].Timeout)*time.Second)) {
		l.Debug().Msg("Timed out")
		return true
	}
	return false
}

func contains(s []string, str string) bool {
	for i := range s {
		if s[i] == str {
			return true
		}
	}
	return false
}
