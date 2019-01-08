package sequencer

import (
	"github.com/geliar/manopus/pkg/matcher"
	"github.com/geliar/manopus/pkg/processor"
)

// SequenceConfig contains description of the execution sequence
type SequenceConfig struct {
	//Name (optional) name of the sequence
	Name string `yaml:"name"`
	//Parallel (optional) instances of this sequence could be executed in parallel
	Parallel bool `yaml:"parallel"`
	//Steps execution steps of the sequence
	Steps []StepConfig `yaml:"steps"`
}

// StepConfig contains description of the sequence step
type StepConfig struct {
	//Name (optional) of the step
	Name string `yaml:"name"`
	//Inputs list of inputs to match
	Inputs []string `yaml:"inputs"`
	//MatchConfig contains matchers
	Match []matcher.MatchConfig `yaml:"match"`
	//Timeout (optional) time (in seconds) to cancel sequence if step is waiting longer
	Timeout int64 `yaml:"timeout"`
	//Processors list of processors to run on event
	Processors []processor.ProcessorConfig `yaml:"processors"`
}
