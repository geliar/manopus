package config

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) {
	// Create a temporary directory for test config files
	dir, err := ioutil.TempDir("", "manopus-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create a dummy YAML config file
	configContent := []byte(`
shutdown_timeout: 10
http:
  listen: ":8080"
report:
  driver: "fs"
  config:
    path: "/tmp/reports"
stores:
  my-store:
    type: "boltdb"
    config:
      path: "/tmp/test.db"
connectors:
  my-connector:
    type: "github"
    config:
      token: "my-secret-token"
sequencer:
  sequences:
    - name: "my-sequence"
      steps:
        - name: "my-step"
          action: "my-action"
`)
	configFile := filepath.Join(dir, "config.yaml")
	if err := ioutil.WriteFile(configFile, configContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Call InitConfig with the dummy config file
	ctx := context.Background()
	config, sequencer, httpServer := InitConfig(ctx, []string{configFile}, false)

	// Assertions
	assert.NotNil(t, config)
	assert.NotNil(t, sequencer)
	assert.NotNil(t, httpServer)
	assert.Equal(t, 10, config.ShutdownTimeout)
	assert.Equal(t, ":8080", config.HTTP.Listen)
	assert.Equal(t, "fs", config.Report.Driver)
	assert.Contains(t, config.Stores, "my-store")
	assert.Equal(t, "boltdb", config.Stores["my-store"].Type)
	assert.Contains(t, config.Connectors, "my-connector")
	assert.Equal(t, "github", config.Connectors["my-connector"].Type)
	assert.NotNil(t, config.Sequencer)
	assert.Len(t, config.Sequencer.SequenceConfigs, 1)
}
