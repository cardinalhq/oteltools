// Copyright 2024-2025 CardinalHQ, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ottl

import (
	"context"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottldatapoint"
	"hash/fnv"
	"slices"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlspan"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

const (
	MetricName = "metricName"
	MetricType = "metricType"
)

type LogExtractor struct {
	Conditions          []*ottl.Condition[ottllog.TransformContext]
	LineDimensions      map[int64]map[string]*ottl.Statement[ottllog.TransformContext]
	AggregateDimensions map[string]*ottl.Statement[ottllog.TransformContext]
	MetricName          string
	RuleID              string
	MetricUnit          string
	MetricType          string
	MetricValue         *ottl.Statement[ottllog.TransformContext]
}

func (l LogExtractor) ExtractLineAttributes(ctx context.Context, tCtx ottllog.TransformContext) map[int64]map[string]any {
	attributesByTagFamilyID := make(map[int64]map[string]any, 0)
	for tagFamilyID, dimensions := range l.LineDimensions {
		attributesByTagFamilyID[tagFamilyID] = l.extractAttributes(ctx, tCtx, dimensions)
	}
	return attributesByTagFamilyID
}

func (l LogExtractor) ExtractAggregateAttributes(ctx context.Context, tCtx ottllog.TransformContext) map[string]any {
	if len(l.AggregateDimensions) == 0 {
		mapAttrs := make(map[string]any)
		mapAttrs[MetricName] = l.MetricName
		mapAttrs[MetricType] = l.MetricType
	}
	return l.extractAttributes(ctx, tCtx, l.AggregateDimensions)
}

func (l LogExtractor) extractAttributes(ctx context.Context, tCtx ottllog.TransformContext, dims map[string]*ottl.Statement[ottllog.TransformContext]) map[string]any {
	if dims == nil {
		return nil
	}

	attrMap := make(map[string]any, len(dims))
	for k, v := range dims {
		attrVal, _, err := v.Execute(ctx, tCtx)
		if err != nil || attrVal == nil || attrVal == "" {
			continue
		}
		attrMap[k] = attrVal
	}
	return attrMap
}

func parseLogExtractorConfig(extractorConfig MetricExtractorConfig, parser ottl.Parser[ottllog.TransformContext]) (*LogExtractor, error) {
	conditions, err := parser.ParseConditions(extractorConfig.Conditions)
	if err != nil {
		return nil, err
	}
	lineDimensionsByTagFamilyID := make(map[int64]map[string]*ottl.Statement[ottllog.TransformContext], 0)
	for _, dim := range extractorConfig.LineDimensions {
		lineDimensions := make(map[string]*ottl.Statement[ottllog.TransformContext])
		tagNames := make([]string, 0)
		for key, value := range dim {
			tagNames = append(tagNames, key)
			statement, statementParseError := parser.ParseStatement(valueStatement(value))
			if statementParseError != nil {
				return nil, statementParseError
			}
			lineDimensions[key] = statement
		}
		slices.Sort(tagNames)
		h := fnv.New64a()
		for _, k := range tagNames {
			_, _ = h.Write([]byte(k))
		}
		tagFamilyID := int64(h.Sum64())
		lineDimensionsByTagFamilyID[tagFamilyID] = lineDimensions
	}

	aggregateDimensions := make(map[string]*ottl.Statement[ottllog.TransformContext])
	for key, value := range extractorConfig.AggregateDimensions {
		statement, statementParseError := parser.ParseStatement(valueStatement(value))
		if statementParseError != nil {
			return nil, statementParseError
		}
		aggregateDimensions[key] = statement
	}
	metricValue, _ := parser.ParseStatement(valueStatement(extractorConfig.MetricValue))

	return &LogExtractor{
		RuleID:              extractorConfig.RuleId,
		Conditions:          conditions,
		LineDimensions:      lineDimensionsByTagFamilyID,
		AggregateDimensions: aggregateDimensions,
		MetricName:          extractorConfig.MetricName,
		MetricUnit:          extractorConfig.MetricUnit,
		MetricType:          extractorConfig.MetricType,
		MetricValue:         metricValue,
	}, nil
}

func (s SpanExtractor) extractAttributes(ctx context.Context, tCtx ottlspan.TransformContext, dims map[string]*ottl.Statement[ottlspan.TransformContext]) map[string]any {
	if dims == nil {
		return nil
	}

	attrMap := make(map[string]any, len(dims))
	for k, v := range dims {
		attrVal, _, err := v.Execute(ctx, tCtx)
		if err != nil || attrVal == nil || attrVal == "" {
			continue
		}
		attrMap[k] = attrVal
	}
	return attrMap
}

func (s SpanExtractor) ExtractLineAttributes(ctx context.Context, tCtx ottlspan.TransformContext) map[int64]map[string]any {
	attributesByTagFamilyID := make(map[int64]map[string]any)
	for tagFamilyID, dimensions := range s.LineDimensions {
		attributesByTagFamilyID[tagFamilyID] = s.extractAttributes(ctx, tCtx, dimensions)
	}
	return attributesByTagFamilyID
}

func (s SpanExtractor) ExtractAggregateAttributes(ctx context.Context, tCtx ottlspan.TransformContext) map[string]any {
	if len(s.AggregateDimensions) == 0 {
		mapAttrs := make(map[string]any)
		mapAttrs[MetricName] = s.MetricName
		mapAttrs[MetricType] = s.MetricType
	}
	return s.extractAttributes(ctx, tCtx, s.AggregateDimensions)
}

type SpanExtractor struct {
	RuleID              string
	Conditions          []*ottl.Condition[ottlspan.TransformContext]
	LineDimensions      map[int64]map[string]*ottl.Statement[ottlspan.TransformContext]
	AggregateDimensions map[string]*ottl.Statement[ottlspan.TransformContext]
	MetricName          string
	MetricUnit          string
	MetricType          string
	MetricValue         *ottl.Statement[ottlspan.TransformContext]
}

type MetricSketchExtractor struct {
	RuleID              string
	MetricName          string
	MetricType          string
	Conditions          []*ottl.Condition[ottldatapoint.TransformContext]
	MetricUnit          string
	LineDimensions      map[int64]map[string]*ottl.Statement[ottldatapoint.TransformContext]
	AggregateDimensions map[string]*ottl.Statement[ottldatapoint.TransformContext]
	OutputMetricName    string
}

func (l LogExtractor) EvalLogConditions(ctx context.Context, transformCtx ottllog.TransformContext) (bool, error) {
	for _, condition := range l.Conditions {
		matches, err := condition.Eval(ctx, transformCtx)
		if err != nil {
			return false, err
		}
		if !matches {
			return false, nil
		}
	}
	return true, nil
}

func (s SpanExtractor) EvalSpanConditions(ctx context.Context, transformCtx ottlspan.TransformContext) (bool, error) {
	for _, condition := range s.Conditions {
		matches, err := condition.Eval(ctx, transformCtx)
		if err != nil {
			return false, err
		}
		if !matches {
			return false, nil
		}
	}
	return true, nil
}

func parseSpanExtractorConfig(
	extractorConfig MetricExtractorConfig,
	parser ottl.Parser[ottlspan.TransformContext],
) (*SpanExtractor, error) {
	conditions, err := parser.ParseConditions(extractorConfig.Conditions)
	if err != nil {
		return nil, err
	}

	lineDimensionsByTagFamilyID := make(map[int64]map[string]*ottl.Statement[ottlspan.TransformContext], 0)
	for _, dim := range extractorConfig.LineDimensions {
		lineDimensions := make(map[string]*ottl.Statement[ottlspan.TransformContext])
		tagNames := make([]string, 0)

		for key, value := range dim {
			tagNames = append(tagNames, key)
			statement, err := parser.ParseStatement(valueStatement(value))
			if err != nil {
				return nil, err
			}
			lineDimensions[key] = statement
		}

		slices.Sort(tagNames)
		h := fnv.New64a()
		for _, k := range tagNames {
			_, _ = h.Write([]byte(k))
		}
		tagFamilyID := int64(h.Sum64())
		lineDimensionsByTagFamilyID[tagFamilyID] = lineDimensions
	}

	aggregateDimensions := make(map[string]*ottl.Statement[ottlspan.TransformContext])
	for key, value := range extractorConfig.AggregateDimensions {
		statement, err := parser.ParseStatement(valueStatement(value))
		if err != nil {
			return nil, err
		}
		aggregateDimensions[key] = statement
	}

	metricValue, err := parser.ParseStatement(valueStatement(extractorConfig.MetricValue))
	if err != nil {
		return nil, err
	}

	return &SpanExtractor{
		RuleID:              extractorConfig.RuleId,
		Conditions:          conditions,
		LineDimensions:      lineDimensionsByTagFamilyID,
		AggregateDimensions: aggregateDimensions,
		MetricName:          extractorConfig.MetricName,
		MetricUnit:          extractorConfig.MetricUnit,
		MetricType:          extractorConfig.MetricType,
		MetricValue:         metricValue,
	}, nil
}

func ParseLogExtractorConfigs(extractorConfigs []MetricExtractorConfig, logger *zap.Logger) ([]*LogExtractor, error) {
	logParser, _ := ottllog.NewParser(ToFactory[ottllog.TransformContext](), component.TelemetrySettings{Logger: logger})

	var logExtractors []*LogExtractor
	for _, extractorConfig := range extractorConfigs {
		logExtractor, err := parseLogExtractorConfig(extractorConfig, logParser)
		if err != nil {
			return nil, err
		}
		logExtractors = append(logExtractors, logExtractor)
	}
	return logExtractors, nil
}

func ParseSpanExtractorConfigs(extractorConfigs []MetricExtractorConfig, logger *zap.Logger) ([]*SpanExtractor, error) {
	spanParser, _ := ottlspan.NewParser(ToFactory[ottlspan.TransformContext](), component.TelemetrySettings{Logger: logger})

	var spanExtractors []*SpanExtractor
	for _, extractorConfig := range extractorConfigs {
		spanExtractor, err := parseSpanExtractorConfig(extractorConfig, spanParser)
		if err != nil {
			return nil, err
		}
		spanExtractors = append(spanExtractors, spanExtractor)
	}
	return spanExtractors, nil
}

func ParseMetricSketchExtractorConfigs(extractorConfigs []MetricSketchExtractorConfig, logger *zap.Logger) (map[string]*MetricSketchExtractor, error) {
	configsByMetricName := make(map[string]*MetricSketchExtractor, len(extractorConfigs))
	parser, _ := ottldatapoint.NewParser(ToFactory[ottldatapoint.TransformContext](), component.TelemetrySettings{Logger: logger})

	for _, extractorConfig := range extractorConfigs {
		conditions, err := parser.ParseConditions(extractorConfig.Conditions)
		if err != nil {
			return nil, err
		}

		m := &MetricSketchExtractor{
			RuleID:           extractorConfig.RuleId,
			Conditions:       conditions,
			MetricName:       extractorConfig.MetricName,
			MetricType:       extractorConfig.MetricType,
			OutputMetricName: extractorConfig.OutputMetricName,
			MetricUnit:       extractorConfig.MetricUnit,
		}
		lineDimensionsByTagFamilyID := make(map[int64]map[string]*ottl.Statement[ottldatapoint.TransformContext])
		for _, dim := range extractorConfig.LineDimensions {
			tagNames := make([]string, 0)
			lineDimensions := make(map[string]*ottl.Statement[ottldatapoint.TransformContext])
			for key, value := range dim {
				tagNames = append(tagNames, key)
				statement, statementParseError := parser.ParseStatement(valueStatement(value))
				if statementParseError != nil {
					return nil, statementParseError
				}
				lineDimensions[key] = statement
			}
			slices.Sort(tagNames)
			h := fnv.New64a()
			for _, k := range tagNames {
				_, _ = h.Write([]byte(k))
			}
			tagFamilyID := int64(h.Sum64())
			lineDimensionsByTagFamilyID[tagFamilyID] = lineDimensions
		}
		aggregateDimensions := make(map[string]*ottl.Statement[ottldatapoint.TransformContext])
		for key, value := range extractorConfig.AggregateDimensions {
			statement, statementParseError := parser.ParseStatement(valueStatement(value))
			if statementParseError != nil {
				return nil, statementParseError
			}
			aggregateDimensions[key] = statement
		}
		configsByMetricName[extractorConfig.MetricName] = m
	}
	return configsByMetricName, nil
}

func (m MetricSketchExtractor) ExtractLineAttributes(ctx context.Context, tCtx ottldatapoint.TransformContext) map[int64]map[string]any {
	attributesByTagFamilyID := make(map[int64]map[string]any)
	for tagFamilyID, dimensions := range m.LineDimensions {
		attributesByTagFamilyID[tagFamilyID] = m.extractAttributes(ctx, tCtx, dimensions)
	}
	return attributesByTagFamilyID
}

func (m MetricSketchExtractor) ExtractAggregateAttributes(ctx context.Context, tCtx ottldatapoint.TransformContext) map[string]any {
	if len(m.AggregateDimensions) == 0 {
		mapAttrs := make(map[string]any)
		mapAttrs[MetricName] = m.MetricName
		mapAttrs[MetricType] = m.MetricType
	}
	return m.extractAttributes(ctx, tCtx, m.AggregateDimensions)
}

func (m MetricSketchExtractor) extractAttributes(ctx context.Context, tCtx ottldatapoint.TransformContext, dims map[string]*ottl.Statement[ottldatapoint.TransformContext]) map[string]any {
	if dims == nil {
		return nil
	}

	attrMap := make(map[string]any, len(dims))
	for k, v := range dims {
		attrVal, _, err := v.Execute(ctx, tCtx)
		if err != nil || attrVal == nil || attrVal == "" {
			continue
		}
		attrMap[k] = attrVal
	}
	return attrMap
}

// This is roughly 5x faster than using fmt.Sprintf()
func valueStatement(value string) string {
	return "value(" + value + ")"
}
