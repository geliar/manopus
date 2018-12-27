package sequencer

import (
	"context"
	"sync"

	"github.com/davecgh/go-spew/spew"

	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/payload"
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
		s.pushnew(sc)
	}
}

func (s *Sequencer) Roll(ctx context.Context, event *input.Event) {
	l := logger(ctx).With().
		Str("event_input", event.Input).
		Str("event_type", event.Type).
		Logger()

	gclist := s.queue.GC(ctx)
	for _, seq := range gclist {
		s.pushnew(seq.sequenceConfig)
		l.Debug().Str("sequence_name", seq.sequenceConfig.Name).Msg("Cleaning timed out sequence")
	}

	sequences := s.queue.Match(ctx, s.DefaultInputs, event)
	for _, seq := range sequences {
		l = l.With().
			Str("sequence_name", seq.sequenceConfig.Name).
			Int("sequence_step", seq.step).
			Logger()
		ctx = l.WithContext(ctx)
		l.Debug().
			Msg("Event matched")
		spew.Dump(seq.payload)
		if seq.step < len(seq.sequenceConfig.Steps)-1 {
			for _, e := range seq.sequenceConfig.Steps[seq.step].Export {
				seq.payload.ExportField(ctx, e.Current, e.New)
				l.Debug().
					Str("export_current", e.Current).
					Str("export_new", e.New).
					Msg("Exported variable")
			}
			seq.step++
			s.queue.Push(seq)
			l.Debug().
				Msg("Next step")
		} else {
			s.pushnew(seq.sequenceConfig)
			l.Debug().Msg("Sequence is finished. Restarting.")
		}
	}
}

func (s *Sequencer) pushnew(sc SequenceConfig) {
	seq := &Sequence{
		sequenceConfig: sc,
		payload:        &payload.Payload{Env: s.Env},
		step:           0,
	}
	s.queue.PushIfNotExists(seq)
}
