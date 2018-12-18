package sequencer

import (
	"sync"
)

type contextElement struct {
	previous *contextElement
	next     *contextElement
	sequence *Sequence
}

type sequenceStack struct {
	first *contextElement
	last  *contextElement
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
	if s.last == nil {
		s.last = s.first
	}
}

func (s *sequenceStack) Pop() (seq *Sequence) {
	s.Lock()
	defer s.Unlock()
	if s.last == nil {
		return nil
	}
	seq = s.last.sequence
	if s.first == s.last {
		s.first = nil
		s.last = nil
		return
	}
	s.last = s.last.previous
	s.last.next = nil
	return
}
