package sequencer

import (
	"context"

	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/payload"
)

type Sequence struct {
	handler HandlerConfig
	step    int
	payload *payload.Payload
}

func (s *Sequence) Match(ctx context.Context, event input.Event) {
	newpayload := *(s.payload)
	newpayload.Req = event.Data
	s.handler.Steps[s.step].Match[0].Match(ctx, &newpayload)
	*(s.payload) = newpayload
}
