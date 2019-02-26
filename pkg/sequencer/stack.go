package sequencer

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"

	"github.com/geliar/manopus/pkg/payload"
)

type contextElement struct {
	previous *contextElement
	next     *contextElement
	sequence *Sequence
}

type sequenceStack struct {
	first *contextElement
	sync.RWMutex
}

func (s *sequenceStack) Push(sequence *Sequence) {
	s.Lock()
	defer s.Unlock()
	s.push(sequence)
}

func (s *sequenceStack) PushIfNotExists(sequence *Sequence) {
	s.Lock()
	defer s.Unlock()
	if !s.exists(sequence) {
		s.push(sequence)
	}
}

func (s *sequenceStack) Exists(sequence *Sequence) bool {
	s.RLock()
	defer s.RUnlock()
	return s.exists(sequence)
}

// Match matching event with sequences in stack, pops and returns first matched sequence
func (s *sequenceStack) Match(ctx context.Context, inputs []string, processorName string, event *payload.Event) (sequences []*Sequence) {
	s.Lock()
	defer s.Unlock()
	elem := s.first
	for elem != nil {
		if elem.sequence.Match(ctx, inputs, processorName, event) {
			s.pop(elem)
			sequences = append(sequences, elem.sequence)
		}
		elem = elem.next
	}
	return
}

func (s *sequenceStack) GC(ctx context.Context) (sequences []*Sequence) {
	s.Lock()
	defer s.Unlock()
	elem := s.first
	for elem != nil {
		if elem.sequence.TimedOut(ctx) {
			s.pop(elem)
			sequences = append(sequences, elem.sequence)
		}
		elem = elem.next
	}
	return
}

func (s *sequenceStack) Len(ctx context.Context) (len int) {
	s.RLock()
	defer s.RUnlock()
	elem := s.first
	for elem != nil {
		len++
		elem = elem.next
	}
	return
}

// pop removes element from stack.
// Warning: pop is not thread-safe sequenceStack should be locked before use
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

// push adds element to the beginning of the stack.
// Warning: push is not thread-safe sequenceStack should be locked before use
func (s *sequenceStack) push(sequence *Sequence) {
	e := &contextElement{next: s.first, sequence: sequence}
	if s.first != nil {
		s.first.previous = e
	}
	s.first = e
}

// exists checks existence of the element in the stack.
// Warning: exists is not thread-safe sequenceStack should be locked before use
func (s *sequenceStack) exists(sequence *Sequence) bool {
	elem := s.first
	for elem != nil {
		if reflect.DeepEqual(elem.sequence.sequenceConfig, sequence.sequenceConfig) &&
			elem.sequence.step == sequence.step {
			return true
		}
		elem = elem.next
	}
	return false
}

func (s *sequenceStack) MarshalJSON() ([]byte, error) {
	s.Lock()
	defer s.Unlock()
	var v []Sequence
	elem := s.first
	for elem != nil {
		if elem.sequence.step != 0 {
			v = append(v, *(elem.sequence))
		}
		elem = elem.next
	}
	return json.Marshal(v)
}

func (s *sequenceStack) UnmarshalJSON(buf []byte) (err error) {
	s.Lock()
	defer s.Unlock()
	var v []Sequence

	err = json.Unmarshal(buf, &v)
	if err != nil {
		return
	}
	for _, e := range v {
		s.push(&e)
	}
	return
}
