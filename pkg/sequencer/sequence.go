package sequencer

import (
	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/payload"
)

type Sequence struct {
	handler HandlerConfig
	step    int
	payload *payload.Payload
}

func (s *Sequence) Match(event input.Event) {
	newpayload := *(s.payload)
	newpayload.Req = event.Data
	s.handler.Steps[s.step].Match[0].Match(&newpayload)
	*(s.payload) = newpayload
}
