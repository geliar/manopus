// +build integration

package config

import (
	"context"
	"testing"

	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/output"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) {
	a := assert.New(t)
	l := log.Logger
	ctx := l.WithContext(context.Background())

	c, s, h := InitConfig(ctx, []string{"../../examples/test"})
	a.NotNil(c)
	a.NotNil(s)
	a.NotNil(h)
	s.Stop(ctx)
	h.Stop(ctx)
	input.StopAll(ctx)
	output.StopAll(ctx)
}
