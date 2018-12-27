package sequencer

import (
	"context"
	"sync"

	"github.com/geliar/manopus/pkg/payload"

	"github.com/geliar/manopus/pkg/input"
)

type Sequencer struct {
	//Env variables which represent env part of context data
	Env map[string]interface{} `yaml:"env"`
	//DefaultInputs the list of inputs which should be matched if inputs list in step if empty
	DefaultInputs []string `yaml:"default_inputs"`
	//SequenceConfigs the list of sequence configs
	SequenceConfigs []SequenceConfig `yaml:"sequences"`
	queue           sequenceStack
	sync.RWMutex
}

func (s *Sequencer) Init(ctx context.Context) {
	for _, sc := range s.SequenceConfigs {
		s.push(sc)
	}
}

func (s *Sequencer) Roll(ctx context.Context, event *input.Event) {
	l := logger(ctx).With().
		Str("event_input", event.Input).
		Str("event_type", event.Type).
		Logger()
	_ = l
	seq := s.queue.Match(ctx, s.DefaultInputs, event)
	// Skip event if not matched
	if seq == nil {
		return
	}
}

func (s *Sequencer) push(sc SequenceConfig) {
	s.queue.Push(&Sequence{
		sequenceConfig: sc,
		payload:        &payload.Payload{Env: s.Env},
		step:           0,
	})
}
