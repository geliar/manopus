package sequencer

import "github.com/geliar/manopus/pkg/matcher"

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
	//Type of the step executor
	Type string `yaml:"type"`
	//Inputs list of inputs to match
	Inputs []string `yaml:"inputs"`
	//MatchConfig contains matchers
	Match []matcher.MatchConfig `yaml:"match"`
	//Timeout (optional) time (in seconds) to cancel sequence if step is waiting longer
	Timeout int64 `yaml:"timeout"`
	//MaxExecutionTime (optional) maximum time (in seconds) a step can execute for
	MaxExecutionTime int64 `yaml:"max_execution_time"`
	//Export list of variables to be exported after execution of step
	Export []struct {
		//Current variable name in payload
		Current string `yaml:"current"`
		//New variable name in export part of payload
		New string `yaml:"new"`
	} `yaml:"export"`
}
