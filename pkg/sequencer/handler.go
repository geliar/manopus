package sequencer

import "github.com/geliar/manopus/pkg/matcher"

// HandlerConfig contains description of the execution sequence
type HandlerConfig struct {
	//Name (optional) name of the handler
	Name string `yaml:"name"`
	//Parallel (optional) instances of this sequence could be executed in parallel
	Parallel bool `yaml:"parallel"`
	//Steps execution steps of the handler
	Steps []StepConfig `yaml:"steps"`
}

// StepConfig contains description of the sequence step
type StepConfig struct {
	//Name (optional) of the step
	Name string `yaml:"name"`
	//Type of the step executor
	Type string `yaml:"type"`
	//MatchConfig contains matchers
	Match []matcher.MatchConfig `yaml:"match"`
	//Timeout (optional) time (in minutes) to cancel sequence if step is lasting longer
	Timeout int64 `yaml:"timeout"`
	//MaxTime (optional) maximum time (in minutes) a step can execute for.
	MaxTime int64 `yaml:"maxtime"`
}