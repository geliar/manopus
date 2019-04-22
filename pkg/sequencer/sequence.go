package sequencer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/geliar/manopus/pkg/payload"
	"github.com/geliar/manopus/pkg/processor"
	"github.com/geliar/manopus/pkg/report"
)

type sequence struct {
	id             string
	sequenceConfig SequenceConfig
	step           int
	event          *payload.Event
	payload        *payload.Payload
	latestMatch    time.Time
}

func (s *sequence) Match(ctx context.Context, inputs []string, processorName string, event *payload.Event) (matched bool) {
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
	} else if len(s.sequenceConfig.Inputs) > 0 {
		if !contains(s.sequenceConfig.Inputs, event.Input) {
			return false
		}
	} else if !contains(inputs, event.Input) {
		return false
	}

	if len(step.Types) > 0 {
		if !contains(step.Types, event.Type) {
			return false
		}
	}

	newPayload := *(s.payload)
	newPayload.Vars = step.Vars
	newPayload.Req = event.Data
	newPayload.Event = new(payload.EventInfo)
	newPayload.Event.Type = event.Type
	newPayload.Event.Input = event.Input
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
	s.latestMatch = time.Now().UTC()
	return true
}

func (s *sequence) Run(ctx context.Context, reporter report.Driver, processorName string) (next processor.NextStatus, callback interface{}, responses []payload.Response) {
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
		next, callback, responses, _ = processor.Run(runCtx, reporter, processorName, step.Script, s.event, &newPayload)
		*(s.payload) = newPayload
		s.latestMatch = time.Now().UTC()
		return
	}
	l.Warn().Msg("script field is empty for the step, there is nothing to execute")
	return
}

func (s *sequence) TimedOut(ctx context.Context) bool {
	l := logger(ctx)
	l = l.With().
		Str("sequence_name", s.sequenceConfig.Name).
		Int("sequence_step", s.step).Logger()
	if s.sequenceConfig.Steps[s.step].Timeout > 0 && time.Now().UTC().After(s.latestMatch.Add(time.Duration(s.sequenceConfig.Steps[s.step].Timeout)*time.Second)) {
		l.Debug().Msg("Timed out")
		return true
	}
	return false
}

func (s *sequence) MarshalJSON() ([]byte, error) {
	compat := struct {
		SequenceConfig SequenceConfig
		Step           int
		ID             string
		Env            map[string]interface{}
		Export         map[string]interface{}
		LatestMatch    int64
	}{
		SequenceConfig: s.sequenceConfig,
		Step:           s.step,
		ID:             s.id,
		Env:            s.payload.Env,
		Export:         s.payload.Export,
		LatestMatch:    s.latestMatch.Unix(),
	}
	return json.Marshal(compat)
}

func (s *sequence) UnmarshalJSON(buf []byte) (err error) {
	compat := struct {
		SequenceConfig SequenceConfig
		Step           int
		ID             string
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
	s.id = compat.ID
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
