package ottl

import (
	"testing"

	"encoding/json"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestControlPlaneConfigYAML(t *testing.T) {
	yamlData := `
pitbulls:
  example:
    log_statements:
      - rule_id: "exampleRule1"
`
	var config ControlPlaneConfig
	err := yaml.Unmarshal([]byte(yamlData), &config)
	assert.NoError(t, err)
	assert.NotNil(t, config.Pitbulls["example"])
}

func TestControlPlaneConfigJSON(t *testing.T) {
	jsonData := `{
	"pitbulls": {
		"example": {
			"log_statements": [
				{
					"rule_id": "exampleRule1"
				}
			]
		}
	}
}`
	var config ControlPlaneConfig
	err := json.Unmarshal([]byte(jsonData), &config)
	assert.NoError(t, err)
	assert.NotNil(t, config.Pitbulls["example"])
}
