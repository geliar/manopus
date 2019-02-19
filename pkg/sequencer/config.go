package sequencer

// SequenceConfig contains description of the execution sequence
type SequenceConfig struct {
	//Name (optional) name of the sequence
	Name string `yaml:"name" json:"name"`
	//Parallel (optional) instances of this sequence could be executed in parallel
	Parallel bool `yaml:"parallel" json:"parallel"`
	//Steps execution steps of the sequence
	Steps []StepConfig `yaml:"steps" json:"steps"`
	//Processor name of processor to run the script
	Processor string `yaml:"processor" json:"processor"`
}

// StepConfig contains description of the sequence step
type StepConfig struct {
	//Name (optional) of the step
	Name string `yaml:"name" json:"name"`
	//Inputs list of inputs to match
	Inputs []string `yaml:"inputs" json:"inputs"`
	//Vars list of variables to be added to payload vars field
	Vars map[string]interface{} `yaml:"vars" json:"vars"`
	//Match contains matcher script
	Match interface{} `yaml:"match" json:"match"`
	//Script contains script to execute on successful match
	Script interface{} `yaml:"script" json:"script"`
	//Timeout (optional) time (in seconds) to cancel sequence if step is waiting longer
	Timeout int64 `yaml:"timeout" json:"timeout"`
	//MaxExecutionTime (optional) maximum time (in seconds) of execution of the script
	MaxExecutionTime int64 `yaml:"max_execution_time" json:"max_execution_time"`
	//Processor name of processor to run the script
	Processor string `yaml:"processor" json:"processor"`
}
