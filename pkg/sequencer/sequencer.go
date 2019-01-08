package sequencer

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/geliar/manopus/pkg/output"
	"github.com/geliar/manopus/pkg/payload"
	"github.com/geliar/manopus/pkg/processor"
)

type Sequencer struct {
	//Env variables which represent env part of context data
	Env map[string]interface{} `yaml:"env"`
	//DefaultInputs the list of inputs which should be matched if inputs list in step is empty
	DefaultInputs []string `yaml:"default_inputs"`
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

func (s *Sequencer) Roll(ctx context.Context, event *payload.Event) (response *payload.Response) {
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

	sequences := s.queue.Match(ctx, s.DefaultInputs, event)
	for _, seq := range sequences {
		l = l.With().
			Str("sequence_name", seq.sequenceConfig.Name).
			Int("sequence_step", seq.step).
			Logger()
		ctx = l.WithContext(ctx)
		l.Debug().
			Msg("Event matched")
		var next processor.NextStatus
		// Running specified processor
		for _, pc := range seq.sequenceConfig.Steps[seq.step].Processors {
			if pc.Type != "" && pc.Script != nil {
				var res interface{}
				runCtx := ctx
				var cancel context.CancelFunc
				if pc.MaxExecutionTime != 0 {
					runCtx, cancel = context.WithTimeout(runCtx, time.Duration(pc.MaxExecutionTime)*time.Second)
				}
				res, next, _ = processor.Run(runCtx, &pc, seq.payload)
				if cancel != nil {
					cancel()
				}

				//Sending request to output
				resp := s.prepareResponse(ctx, pc.Encoding, event, res)
				if pc.Destination != "" {
					s.sendToOutput(ctx, pc.Destination, resp)
				}
				if resp != nil {
					if !pc.SkipCallback {
						response = resp
					}
					seq.payload.Resp = resp.Data
				}
				for _, e := range pc.Export {
					seq.payload.ExportField(ctx, e.Current, e.New)
					l.Debug().
						Str("export_current", e.Current).
						Str("export_new", e.New).
						Msg("Exported variable")
				}
				if next == processor.NextStopSequence {
					break
				}
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
			l.Debug().Msg("Sequence is finished. Creating new.")
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

func (s *Sequencer) prepareResponse(ctx context.Context, encoding string, req *payload.Event, res interface{}) (response *payload.Response) {
	l := logger(ctx)
	// Decoding response
	var resp map[string]interface{}
	switch encoding {
	case "none":
		var ok bool
		if resp, ok = res.(map[string]interface{}); !ok {
			l.Error().
				Msg("Wrong type of response returned from processor")
			return
		}
	case "json", "toml", "plain":
		var buf []byte
		switch v := res.(type) {
		case string:
			buf = []byte(v)
		case []byte:
			buf = v
		default:
			l.Error().
				Msg("Wrong type of response returned from processor")
			return
		}
		switch encoding {
		case "json":
			err := json.Unmarshal(buf, &resp)
			if err != nil {
				l.Warn().
					Err(err).
					Str("response_data", string(buf)).
					Msg("Cannot unmarshal JSON response returned from processor")
				return
			}
		case "toml":
			err := toml.Unmarshal(buf, &resp)
			if err != nil {
				l.Warn().
					Err(err).
					Str("response_data", string(buf)).
					Msg("Cannot unmarshal TOML response returned from processor")
				return
			}
		case "plain":
			resp = map[string]interface{}{
				"data": res,
			}
		}
	case "":
		l.Debug().
			Msg("Encoding is empty. Skipping response.")
		return
	default:
		l.Error().
			Msg("Wrong encoding")
		return
	}

	return &payload.Response{
		ID:       req.ID,
		Request:  req,
		Encoding: encoding,
		Data:     resp,
	}
}

func (s *Sequencer) sendToOutput(ctx context.Context, destination string, response *payload.Response) {
	if response == nil {
		return
	}
	l := logger(ctx).With().Str("output_destination", destination).
		Str("response_encoding", response.Encoding).Logger()
	output.Send(l.WithContext(ctx), destination, response)
}

func (s *Sequencer) pushnew(sc SequenceConfig) {
	seq := &Sequence{
		sequenceConfig: sc,
		payload:        &payload.Payload{Env: s.Env},
		step:           0,
	}
	s.queue.PushIfNotExists(seq)
}
