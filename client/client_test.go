package client

import (
	"encoding/json"
	"testing"

	"github.com/coinbase/bifrost/deployer"
	"github.com/stretchr/testify/assert"
)

func minimalRelease(t *testing.T) *deployer.Release {
	var r deployer.Release
	err := json.Unmarshal([]byte(`
  {
    "release_id": "rr",
    "project_name": "project",
    "config_name": "config",
    "image": "ami-123456"
  }
  `), &r)

	assert.NoError(t, err)
	return &r
}
