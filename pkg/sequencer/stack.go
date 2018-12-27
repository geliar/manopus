package sequencer

import (
	"context"
	"sync"

	"github.com/geliar/manopus/pkg/input"
)

type contextElement struct {
	previous *contextElement
	next     *contextElement
	sequence *Sequence
}

type sequenceStack struct {
	first *contextElement
	sync.Mutex
}

func (s *sequenceStack) Push(sequence *Sequence) {
	s.Lock()
	defer s.Unlock()
	e := &contextElement{next: s.first, sequence: sequence}
	if s.first != nil {
		s.first.previous = e
	}
	s.first = e
}

// Match matching event with sequences in stack, pops and returns first matched sequence
func (s *sequenceStack) Match(ctx context.Context, defaultInputs []string, event *input.Event) *Sequence {
	s.Lock()
	defer s.Unlock()
	elem := s.first
	for elem != nil {
		if elem.sequence.Match(ctx, defaultInputs, event) {
			s.pop(elem)
			return elem.sequence
		}
		elem = elem.next
	}
	return nil
}

// Removing element from stack.
// Warning: pop is not threadsafe sequenceStack should be locked before use
func (s *sequenceStack) pop(elem *contextElement) {
	if elem.next != nil {
		elem.next.previous = elem.previous
	}
	if elem.previous != nil {
		elem.previous.next = elem.next
	}
	if s.first == elem {
		s.first = elem.next
	}
}
