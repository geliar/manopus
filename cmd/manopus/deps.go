package main

import (
	//Connectors
	_ "github.com/geliar/manopus/pkg/connector/http"
	_ "github.com/geliar/manopus/pkg/connector/slack"
	_ "github.com/geliar/manopus/pkg/connector/timer"

	//Processors
	_ "github.com/geliar/manopus/pkg/processor/starlark"
)
