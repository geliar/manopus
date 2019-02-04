package sequencer

import (
	"context"
	"sync"

	"github.com/geliar/manopus/pkg/output"
	"github.com/geliar/manopus/pkg/payload"
	"github.com/geliar/manopus/pkg/processor"
)

type Sequencer struct {
	//Env variables which represent env part of context data
	Env map[string]interface{} `yaml:"env"`
	//Inputs the list of inputs which should be matched if inputs list in step is empty
	Inputs []string `yaml:"inputs"`
	//Processor name of processor to run the scripts
	Processor string `yaml:"processor"`
	//SequenceConfigs the list of sequence configs
	SequenceConfigs []SequenceConfig `yaml:"sequences"`
	queue           sequenceStack
	stop            bool
	sync.RWMutex
}

func (s *Sequencer) Init(ctx context.Context) {
	for _, sc := range s.SequenceConfigs {
		s.pushnew(sc)
	}
}

func (s *Sequencer) Roll(ctx context.Context, event *payload.Event) (response interface{}) {
	l := logger(ctx).With().
		Str("event_input", event.Input).
		Str("event_type", event.Type).
		Str("event_id", event.ID).
		Logger()
	s.RLock()
	defer s.RUnlock()
	if s.stop {
		l.Debug().Msg("Received event on stopped Sequencer")
		return
	}
	gclist := s.queue.GC(ctx)
	for _, seq := range gclist {
		s.pushnew(seq.sequenceConfig)
		l.Debug().Str("sequence_name", seq.sequenceConfig.Name).Msg("Cleaning timed out sequence")
	}

	sequences := s.queue.Match(ctx, s.Inputs, s.Processor, event)
	for _, seq := range sequences {
		if s.stop {
			return
		}
		l = l.With().
			Str("sequence_name", seq.sequenceConfig.Name).
			Int("sequence_step", seq.step).
			Logger()
		ctx = l.WithContext(ctx)
		l.Debug().
			Msg("Event matched")
		var next processor.NextStatus
		// Running specified processor

		next, callback, responses := seq.Run(ctx, s.Processor)

		if callback != nil {
			if response != nil {
				l.Warn().Msg("Multiple sequences returned callback data. Using the latest one.")
			}
			response = callback
		}
		//Sending requests to outputs
		for _, r := range responses {
			if s.stop {
				return
			}
			r.ID = event.ID
			r.Request = event

			if r.Output != "" {
				s.sendToOutput(ctx, &r)
			}
		}

		if s.stop {
			return
		}

		//If this step is not last
		if next == processor.NextRepeatStep ||
			(seq.step < len(seq.sequenceConfig.Steps)-1 &&
				next != processor.NextStopSequence) {
			//Pushing sequence back to queue but with incremented step number
			if next != processor.NextRepeatStep {
				seq.step++
			}
			//Cleanup
			seq.payload.Req = nil
			seq.payload.Resp = nil
			//Pushing sequence back to queue
			s.queue.Push(seq)
			l.Debug().
				Msg("Next step")
		} else {
			//If it is last step starting sequence from beginning
			s.pushnew(seq.sequenceConfig)
			l.Debug().Msg("Sequence is finished. Creating new one.")
		}
	}
	return
}

func (s *Sequencer) Stop(ctx context.Context) {
	l := logger(ctx)
	l.Info().Msg("Shutting down sequencer")
	s.stop = true
	s.Lock()
	defer s.Unlock()
}

func (s *Sequencer) sendToOutput(ctx context.Context, response *payload.Response) {
	if response == nil {
		return
	}
	_ = output.Send(ctx, response)
}

func (s *Sequencer) pushnew(sc SequenceConfig) {
	seq := &Sequence{
		sequenceConfig: sc,
		payload:        &payload.Payload{Env: s.Env},
		step:           0,
	}
	s.queue.PushIfNotExists(seq)
}
