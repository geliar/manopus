package main

import (
	//Connectors
	_ "github.com/geliar/manopus/pkg/connector/http"
	_ "github.com/geliar/manopus/pkg/connector/slack"

	//Processors
	_ "github.com/geliar/manopus/pkg/processor/bash"
	_ "github.com/geliar/manopus/pkg/processor/python"
	_ "github.com/geliar/manopus/pkg/processor/simple"
)
