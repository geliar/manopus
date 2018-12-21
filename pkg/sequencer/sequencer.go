package sequencer

import (
	"context"
	"sync"

	"github.com/geliar/manopus/pkg/config"
	"github.com/geliar/manopus/pkg/input"
)

type Sequencer struct {
	queue *sequenceStack
	sync.RWMutex
}

func (s *Sequencer) AddSequence(sequence config.SequenceConfig) error {
	s.Lock()
	if s.queue == nil {
		s.queue = new(sequenceStack)
	}
	s.Unlock()

	return nil
}

func (s *Sequencer) Roll(ctx context.Context, event input.Event) {
	l := logger(ctx).With().
		Str("event_input", event.Input).
		Str("event_type", event.Type).
		Logger()
	func() {
		s.RLock()
		defer s.RUnlock()
		if s.queue == nil {
			l.Error().Msg("Queue is not initialized. Skipping event.")
			return
		}
	}()
	seq := s.queue.Pop()
	if seq == nil {
		l.Debug().Msg("Empty queue. Skipping event.")
		return
	}
}
