package sequencer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/geliar/manopus/pkg/processor"

	"github.com/geliar/manopus/pkg/payload"
)

type Sequence struct {
	sequenceConfig SequenceConfig
	step           int
	event          *payload.Event
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
	s.event = event
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
		next, callback, responses, _ = processor.Run(runCtx, processorName, step.Script, s.event, &newPayload)
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

func (s *Sequence) MarshalJSON() ([]byte, error) {
	compat := struct {
		SequenceConfig SequenceConfig
		Step           int
		Env            map[string]interface{}
		Export         map[string]interface{}
		LatestMatch    int64
	}{
		SequenceConfig: s.sequenceConfig,
		Step:           s.step,
		Env:            s.payload.Env,
		Export:         s.payload.Export,
		LatestMatch:    s.latestMatch.Unix(),
	}
	return json.Marshal(compat)
}

func (s *Sequence) UnmarshalJSON(buf []byte) (err error) {
	compat := struct {
		SequenceConfig SequenceConfig
		Step           int
		Env            map[string]interface{}
		Export         map[string]interface{}
		LatestMatch    int64
	}{}
	err = json.Unmarshal(buf, &compat)
	if err != nil {
		return
	}
	s.sequenceConfig = compat.SequenceConfig
	s.step = compat.Step
	s.payload = new(payload.Payload)
	s.payload.Env = compat.Env
	s.payload.Export = compat.Export
	s.latestMatch = time.Unix(compat.LatestMatch, 0)
	return
}

func contains(s []string, str string) bool {
	for i := range s {
		if s[i] == str {
			return true
		}
	}
	return false
}
