package sequencer

import "github.com/DLag/manopus/pkg/matcher"

type Handler struct {
	Name string
	Steps []Step
}

type Step struct {
	Name string
	Type string
	Match []matcher.Match
	Timeout int64
}