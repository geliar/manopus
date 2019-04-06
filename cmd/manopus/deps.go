package main

import (
	//Connectors
	_ "github.com/geliar/manopus/pkg/connector/bitbucket"
	_ "github.com/geliar/manopus/pkg/connector/github"
	_ "github.com/geliar/manopus/pkg/connector/http"
	_ "github.com/geliar/manopus/pkg/connector/slack"
	_ "github.com/geliar/manopus/pkg/connector/timer"

	//Processors
	_ "github.com/geliar/manopus/pkg/processor/starlark"

	//Stores
	_ "github.com/geliar/manopus/pkg/store/boltdb"

	//Reporters
	_ "github.com/geliar/manopus/pkg/report/fs"
)
