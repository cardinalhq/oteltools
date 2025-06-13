package ottl

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestParseMetricSketchExtractorConfig(t *testing.T) {
	// Arrange
	filePath := filepath.Join("testdata", "config.json")
	data, err := os.ReadFile(filePath)
	require.NoError(t, err, "failed to read config.json")

	var config MetricSketchExtractorConfig

	err = json.Unmarshal(data, &config)

	// Assert
	require.NoError(t, err, "failed to unmarshal config.json")
	require.NotEmpty(t, config.RuleId, "RuleId should not be empty")
	//require.NotEmpty(t, config.Conditions, "Conditions should not be empty")
	require.NotEmpty(t, config.MetricName, "MetricName should not be empty")
	require.NotZero(t, config.Version, "Version should not be zero")

	//configs, err := ParseMetricSketchExtractorConfigs([]MetricSketchExtractorConfig{config}, nil)
	//require.NoError(t, err, "failed to parse MetricSketchExtractorConfig")
	//require.NotEmpty(t, configs, "Parsed configs should not be empty")
}
