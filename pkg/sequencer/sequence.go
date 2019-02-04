package sequencer

import (
	"context"
	"time"

	"github.com/geliar/manopus/pkg/processor"

	"github.com/geliar/manopus/pkg/payload"
)

type Sequence struct {
	sequenceConfig SequenceConfig
	step           int
	payload        *payload.Payload
	latestMatch    time.Time
}

func (s *Sequence) Match(ctx context.Context, inputs []string, processorName string, event *payload.Event) (matched bool) {
	l := logger(ctx)
	l = l.With().
		Str("sequence_name", s.sequenceConfig.Name).
		Int("sequence_step", s.step).Logger()
	ctx = l.WithContext(ctx)
	step := &s.sequenceConfig.Steps[s.step]
	if len(step.Inputs) > 0 {
		if !contains(step.Inputs, event.Input) {
			return false
		}
	} else {
		if !contains(inputs, event.Input) {
			return false
		}
	}

	newPayload := *(s.payload)
	newPayload.Vars = step.Vars
	newPayload.Req = event.Data
	if step.Match != nil {
		if s.sequenceConfig.Processor != "" {
			processorName = s.sequenceConfig.Processor
		}
		if step.Processor != "" {
			processorName = step.Processor
		}

		matched, _ = processor.Match(ctx, processorName, step.Match, &newPayload)
		if !matched {
			return false
		}
	}
	*(s.payload) = newPayload
	s.latestMatch = time.Now()
	return true
}

func (s *Sequence) Run(ctx context.Context, processorName string) (next processor.NextStatus, callback interface{}, responses []payload.Response) {
	l := logger(ctx)
	l = l.With().
		Str("sequence_name", s.sequenceConfig.Name).
		Int("sequence_step", s.step).Logger()
	ctx = l.WithContext(ctx)
	step := &s.sequenceConfig.Steps[s.step]

	if step.Script != nil {
		runCtx := ctx
		var cancel context.CancelFunc
		if step.MaxExecutionTime != 0 {
			runCtx, cancel = context.WithTimeout(runCtx, time.Duration(step.MaxExecutionTime)*time.Second)
			defer cancel()
		}
		if s.sequenceConfig.Processor != "" {
			processorName = s.sequenceConfig.Processor
		}
		if step.Processor != "" {
			processorName = step.Processor
		}

		newPayload := *(s.payload)
		if newPayload.Export == nil {
			newPayload.Export = make(map[string]interface{})
		}
		next, callback, responses, _ = processor.Run(runCtx, processorName, step.Script, &newPayload)
		*(s.payload) = newPayload
		s.latestMatch = time.Now()
		return
	}
	l.Warn().Msg("script field is empty for the step, there is nothing to execute")
	return
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
