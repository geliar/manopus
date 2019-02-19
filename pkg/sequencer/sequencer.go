package sequencer

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/geliar/manopus/pkg/output"
	"github.com/geliar/manopus/pkg/payload"
	"github.com/geliar/manopus/pkg/processor"
	"github.com/geliar/manopus/pkg/store"
)

type Sequencer struct {
	//Env variables which represent env part of context data
	Env map[string]interface{} `yaml:"env"`
	//Inputs the list of inputs which should be matched if inputs list in step is empty
	Inputs []string `yaml:"inputs"`
	//Processor name of processor to run the scripts
	Processor string `yaml:"processor"`
	//Store name of store to save state of sequencer
	Store string `yaml:"store"`
	//StoreKey key string to use for storing sequencer state
	StoreKey string `yaml:"store_key"`
	//SequenceConfigs the list of sequence configs
	SequenceConfigs []SequenceConfig `yaml:"sequences"`
	queue           sequenceStack
	stop            bool
	sync.RWMutex
	running int64
	mainCtx context.Context
}

func (s *Sequencer) Init(ctx context.Context) {
	s.mainCtx = ctx
	if s.Store != "" && s.StoreKey != "" {
		_ = s.load(ctx)
	}
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
	atomic.AddInt64(&s.running, 1)
	defer atomic.AddInt64(&s.running, -1)
	gclist := s.queue.GC(ctx)
	for _, seq := range gclist {
		s.pushnew(seq.sequenceConfig)
		l.Debug().Str("sequence_name", seq.sequenceConfig.Name).Msg("Cleaning timed out sequence")
	}
	ctx = mergeContexts(s.mainCtx, ctx)
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
		if !seq.sequenceConfig.Single && seq.step == 0 {
			l.Debug().Msg("Sequence can be executed in parallel. Creating new one.")
			s.pushnew(seq.sequenceConfig)
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
			//If it is the last step starting sequence from beginning
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
	if r := atomic.LoadInt64(&s.running); r != 0 {
		l.Info().Msgf("Waiting for %d running sequence(s)", r)
	}
	s.Lock()
	defer s.Unlock()
	_ = s.save(ctx)
}

func (s *Sequencer) load(ctx context.Context) error {
	l := logger(ctx)
	l.Info().Msg("Loading saved sequences from store")
	buf, err := store.Load(ctx, s.Store, s.StoreKey)
	if err != nil {
		l.Error().Err(err).Msg("Error on loading from store")
		return err
	}

	if len(buf) == 0 {
		l.Info().Msg("Sequence store is empty")
		return nil
	}
	var tmp struct {
		Store *sequenceStack
	}
	tmp.Store = &s.queue
	err = json.Unmarshal(buf, &tmp)
	if err != nil {
		l.Error().Err(err).Msg("Error on parsing store value")
		return err
	}
	l.Info().Msgf("Found %d unfinished sequence(s)", s.queue.Len(ctx))
	return nil
}

func (s *Sequencer) save(ctx context.Context) error {
	l := logger(ctx)
	l.Info().Msg("Saving unfinished sequences to store")
	var tmp struct {
		Store *sequenceStack
	}
	tmp.Store = &s.queue
	buf, err := json.Marshal(tmp)
	if err != nil {
		l.Error().Err(err).Msg("Error on dumping sequences to JSON")
		return err
	}
	err = store.Save(ctx, s.Store, s.StoreKey, buf)
	if err != nil {
		l.Error().Err(err).Msg("Error on saving to store")
		return err
	}
	return nil
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
