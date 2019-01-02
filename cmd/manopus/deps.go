package main

import (
	//Connectors
	_ "github.com/geliar/manopus/pkg/connector/slackrtm"
	//Processors
	_ "github.com/geliar/manopus/pkg/processor/bash"
	_ "github.com/geliar/manopus/pkg/processor/simple"
)
