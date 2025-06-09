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

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlspan"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

type LogExtractor struct {
	Conditions       []*ottl.Condition[ottllog.TransformContext]
	MetricDimensions map[string]*ottl.Statement[ottllog.TransformContext]
	SketchDimensions map[string]*ottl.Statement[ottllog.TransformContext]
	MetricName       string
	RuleID           string
	MetricUnit       string
	MetricType       string
	MetricValue      *ottl.Statement[ottllog.TransformContext]
}

func (l LogExtractor) ExtractMetricAttributes(ctx context.Context, tCtx ottllog.TransformContext) map[string]any {
	return l.extractAttributes(ctx, tCtx, l.MetricDimensions)
}

func (l LogExtractor) ExtractSketchAttributes(ctx context.Context, tCtx ottllog.TransformContext) map[string]any {
	return l.extractAttributes(ctx, tCtx, l.SketchDimensions)
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

func (l SpanExtractor) extractAttributes(ctx context.Context, tCtx ottlspan.TransformContext, dims map[string]*ottl.Statement[ottlspan.TransformContext]) map[string]any {
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

func (s SpanExtractor) ExtractMetricAttributes(ctx context.Context, tCtx ottlspan.TransformContext) map[string]any {
	return s.extractAttributes(ctx, tCtx, s.MetricDimensions)
}

func (l SpanExtractor) ExtractSketchAttributes(ctx context.Context, tCtx ottlspan.TransformContext) map[string]any {
	return l.extractAttributes(ctx, tCtx, l.SketchDimensions)
}

type SpanExtractor struct {
	RuleID           string
	Conditions       []*ottl.Condition[ottlspan.TransformContext]
	MetricDimensions map[string]*ottl.Statement[ottlspan.TransformContext]
	SketchDimensions map[string]*ottl.Statement[ottlspan.TransformContext]
	MetricName       string
	MetricUnit       string
	MetricType       string
	MetricValue      *ottl.Statement[ottlspan.TransformContext]
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

func (l SpanExtractor) EvalSpanConditions(ctx context.Context, transformCtx ottlspan.TransformContext) (bool, error) {
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

func parseLogExtractorConfig(extractorConfig MetricExtractorConfig, parser ottl.Parser[ottllog.TransformContext]) (*LogExtractor, error) {
	conditions, err := parser.ParseConditions(extractorConfig.Conditions)
	if err != nil {
		return nil, err
	}
	metricDimensions := make(map[string]*ottl.Statement[ottllog.TransformContext])
	for key, value := range extractorConfig.MetricDimensions {
		statement, statementParseError := parser.ParseStatement(valueStatement(value))
		if statementParseError != nil {
			return nil, statementParseError
		}
		metricDimensions[key] = statement
	}

	sketchDimensions := make(map[string]*ottl.Statement[ottllog.TransformContext])
	for key, value := range extractorConfig.SketchDimensions {
		statement, statementParseError := parser.ParseStatement(valueStatement(value))
		if statementParseError != nil {
			return nil, statementParseError
		}
		metricDimensions[key] = statement
	}
	metricValue, _ := parser.ParseStatement(valueStatement(extractorConfig.MetricValue))

	return &LogExtractor{
		RuleID:           extractorConfig.RuleId,
		Conditions:       conditions,
		MetricDimensions: metricDimensions,
		SketchDimensions: sketchDimensions,
		MetricName:       extractorConfig.MetricName,
		MetricUnit:       extractorConfig.MetricUnit,
		MetricType:       extractorConfig.MetricType,
		MetricValue:      metricValue,
	}, nil
}

func parseSpanExtractorConfig(extractorConfig MetricExtractorConfig, parser ottl.Parser[ottlspan.TransformContext]) (*SpanExtractor, error) {
	conditions, err := parser.ParseConditions(extractorConfig.Conditions)
	if err != nil {
		return nil, err
	}
	metricDimensions := make(map[string]*ottl.Statement[ottlspan.TransformContext])
	for key, value := range extractorConfig.MetricDimensions {
		statement, statementParseError := parser.ParseStatement(valueStatement(value))
		if statementParseError != nil {
			return nil, statementParseError
		}
		metricDimensions[key] = statement
	}
	sketchDimensions := make(map[string]*ottl.Statement[ottlspan.TransformContext])
	for key, value := range extractorConfig.MetricDimensions {
		statement, statementParseError := parser.ParseStatement(valueStatement(value))
		if statementParseError != nil {
			return nil, statementParseError
		}
		sketchDimensions[key] = statement
	}
	metricValue, _ := parser.ParseStatement(valueStatement(extractorConfig.MetricValue))

	return &SpanExtractor{
		RuleID:           extractorConfig.RuleId,
		Conditions:       conditions,
		MetricDimensions: metricDimensions,
		SketchDimensions: sketchDimensions,
		MetricName:       extractorConfig.MetricName,
		MetricUnit:       extractorConfig.MetricUnit,
		MetricType:       extractorConfig.MetricType,
		MetricValue:      metricValue,
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

// This is roughly 5x faster than using fmt.Sprintf()
func valueStatement(value string) string {
	return "value(" + value + ")"
}
