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
	"fmt"
	"math/rand"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottldatapoint"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlresource"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlscope"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlspan"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/ottlfuncs"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/cardinalhq/oteltools/pkg/ottl/functions"
	"github.com/cardinalhq/oteltools/pkg/telemetry"
	"github.com/cardinalhq/oteltools/pkg/translate"
)

type Versioned interface {
	GetVersion() int
}

type resourceTransform struct {
	context    ContextID
	version    int
	conditions []*ottl.Condition[ottlresource.TransformContext]
	statements []*ottl.Statement[ottlresource.TransformContext]
}

func (r *resourceTransform) GetVersion() int {
	return r.version
}

type scopeTransform struct {
	context    ContextID
	version    int
	conditions []*ottl.Condition[ottlscope.TransformContext]
	statements []*ottl.Statement[ottlscope.TransformContext]
}

func (s *scopeTransform) GetVersion() int {
	return s.version
}

type logTransform struct {
	context    ContextID
	version    int
	conditions []*ottl.Condition[ottllog.TransformContext]
	statements []*ottl.Statement[ottllog.TransformContext]
	//samplerConfig SamplingConfig
	sampler Sampler
}

func (s *logTransform) GetVersion() int {
	return s.version
}

type spanTransform struct {
	context    ContextID
	version    int
	conditions []*ottl.Condition[ottlspan.TransformContext]
	statements []*ottl.Statement[ottlspan.TransformContext]
	//samplerConfig SamplingConfig
	sampler Sampler
}

func (s *spanTransform) GetVersion() int {
	return s.version
}

type metricTransform struct {
	context    ContextID
	version    int
	conditions []*ottl.Condition[ottlmetric.TransformContext]
	statements []*ottl.Statement[ottlmetric.TransformContext]
}

func (s *metricTransform) GetVersion() int {
	return s.version
}

type dataPointTransform struct {
	context    ContextID
	version    int
	conditions []*ottl.Condition[ottldatapoint.TransformContext]
	statements []*ottl.Statement[ottldatapoint.TransformContext]
}

func (s *dataPointTransform) GetVersion() int {
	return s.version
}

type RuleID string

type Transformations = transformations

type transformations struct {
	resourceTransforms  map[RuleID]*resourceTransform
	scopeTransforms     map[RuleID]*scopeTransform
	logTransforms       map[RuleID]*logTransform
	spanTransforms      map[RuleID]*spanTransform
	metricTransforms    map[RuleID]*metricTransform
	dataPointTransforms map[RuleID]*dataPointTransform
}

func NewTransformations() *transformations {
	return &transformations{
		resourceTransforms:  make(map[RuleID]*resourceTransform),
		scopeTransforms:     make(map[RuleID]*scopeTransform),
		logTransforms:       make(map[RuleID]*logTransform),
		spanTransforms:      make(map[RuleID]*spanTransform),
		metricTransforms:    make(map[RuleID]*metricTransform),
		dataPointTransforms: make(map[RuleID]*dataPointTransform),
	}
}

func MergeWith(this *transformations, other *transformations) *transformations {
	return &transformations{
		resourceTransforms:  merge(this.resourceTransforms, other.resourceTransforms),
		scopeTransforms:     merge(this.scopeTransforms, other.scopeTransforms),
		logTransforms:       merge(this.logTransforms, other.logTransforms),
		spanTransforms:      merge(this.spanTransforms, other.spanTransforms),
		metricTransforms:    merge(this.metricTransforms, other.metricTransforms),
		dataPointTransforms: merge(this.dataPointTransforms, other.dataPointTransforms),
	}
}

func merge[T any](map1, map2 map[RuleID]T) map[RuleID]T {
	result := make(map[RuleID]T, len(map1))

	for key, value := range map1 {
		result[key] = value
	}
	for key, value := range map2 {
		result[key] = value
	}

	return result
}

func (t *transformations) Stop() {
	for _, logTransform := range t.logTransforms {
		if logTransform.sampler != nil {
			_ = logTransform.sampler.Stop()
		}
	}
	for _, spanTransform := range t.spanTransforms {
		if spanTransform.sampler != nil {
			_ = spanTransform.sampler.Stop()
		}
	}
}

type valueArguments[K any] struct {
	Target ottl.Getter[K] `ottlarg:"0"`
}

func newValueFactory[K any]() ottl.Factory[K] {
	return ottl.NewFactory("value", &valueArguments[K]{}, createValueFunction[K])
}

func createValueFunction[K any](_ ottl.FunctionContext, a ottl.Arguments) (ottl.ExprFunc[K], error) {
	args, ok := a.(*valueArguments[K])
	if !ok {
		return nil, fmt.Errorf("valueFactory args must be of type *valueArguments[K]")
	}

	return valueFn(args)
}

func valueFn[K any](c *valueArguments[K]) (ottl.ExprFunc[K], error) {
	return func(ctx context.Context, tCtx K) (interface{}, error) {
		return c.Target.Get(ctx, tCtx)
	}, nil
}

func ToFactory[T any]() map[string]ottl.Factory[T] {
	factoryMap := map[string]ottl.Factory[T]{}
	for factoryName, factory := range ottlfuncs.StandardFuncs[T]() {
		factoryMap[factoryName] = factory
	}
	for factoryName, factory := range ottlfuncs.StandardConverters[T]() {
		factoryMap[factoryName] = factory
	}
	for factoryName, factory := range functions.CustomFunctions[T]() {
		factoryMap[factoryName] = factory
	}

	valueFactory := newValueFactory[T]()
	factoryMap[valueFactory.Name()] = valueFactory

	return factoryMap
}

func createSampler(c SamplingConfig) Sampler {
	if c.RPS > 0 {
		return NewRPSSampler(WithMaxRPS(c.RPS))
	}
	if c.SampleRate > 0 {
		return NewStaticSampler(int(1 / c.SampleRate))
	}
	return nil
}

func GetServiceName(resource pcommon.Resource) string {
	r := resource.Attributes()
	snk := string(semconv.ServiceNameKey)
	if serviceNameField, found := r.Get(snk); found {
		return serviceNameField.AsString()
	}
	return "unknown"
}

func ParseTransformations(logger *zap.Logger, statements []ContextStatement) (*transformations, error) {
	var errors error

	resourceParser, _ := ottlresource.NewParser(ToFactory[ottlresource.TransformContext](), component.TelemetrySettings{Logger: logger})
	scopeParser, _ := ottlscope.NewParser(ToFactory[ottlscope.TransformContext](), component.TelemetrySettings{Logger: logger})
	logParser, _ := ottllog.NewParser(ToFactory[ottllog.TransformContext](), component.TelemetrySettings{Logger: logger})
	spanParser, _ := ottlspan.NewParser(ToFactory[ottlspan.TransformContext](), component.TelemetrySettings{Logger: logger})
	metricParser, _ := ottlmetric.NewParser(ToFactory[ottlmetric.TransformContext](), component.TelemetrySettings{Logger: logger})
	dataPointParser, _ := ottldatapoint.NewParser(ToFactory[ottldatapoint.TransformContext](), component.TelemetrySettings{Logger: logger})

	transformations := NewTransformations()

	for _, cs := range statements {
		switch cs.Context {
		case "resource":
			conditions, err := resourceParser.ParseConditions(cs.Conditions)
			if err != nil {
				logger.Error("Error parsing resource conditions", zap.Error(err))
				errors = multierr.Append(errors, err)
				continue
			}
			statements, err := resourceParser.ParseStatements(cs.Statements)
			if err != nil {
				logger.Error("Error parsing resource statements", zap.Error(err))
				errors = multierr.Append(errors, err)
				continue
			}

			transformations.resourceTransforms[cs.RuleId] = &resourceTransform{
				context:    cs.Context,
				conditions: conditions,
				statements: statements,
				version:    cs.Version,
			}

		case "scope":
			conditions, err := scopeParser.ParseConditions(cs.Conditions)
			if err != nil {
				logger.Error("Error parsing scope conditions", zap.Error(err))
				errors = multierr.Append(errors, err)
				continue
			}
			statements, err := scopeParser.ParseStatements(cs.Statements)
			if err != nil {
				logger.Error("Error parsing scope statements", zap.Error(err))
				errors = multierr.Append(errors, err)
				continue
			}

			transformations.scopeTransforms[cs.RuleId] = &scopeTransform{
				context:    cs.Context,
				conditions: conditions,
				statements: statements,
				version:    cs.Version,
			}

		case "log":
			conditions, err := logParser.ParseConditions(cs.Conditions)
			if err != nil {
				logger.Error("Error parsing log conditions", zap.Error(err))
				errors = multierr.Append(errors, err)
				continue
			}
			statements, err := logParser.ParseStatements(cs.Statements)
			if err != nil {
				logger.Error("Error parsing log statements", zap.Error(err))
				errors = multierr.Append(errors, err)
				continue
			}

			s := createSampler(cs.SamplingConfig)
			if s != nil {
				err = s.Start()
				if err != nil {
					logger.Error("Error starting sampler", zap.Error(err))
					errors = multierr.Append(errors, err)
					continue
				}
			}

			transformations.logTransforms[cs.RuleId] = &logTransform{
				context:    cs.Context,
				conditions: conditions,
				statements: statements,
				sampler:    s,
				version:    cs.Version,
			}

		case "span":
			conditions, err := spanParser.ParseConditions(cs.Conditions)
			if err != nil {
				logger.Error("Error parsing span conditions", zap.Error(err))
				errors = multierr.Append(errors, err)
				continue
			}
			statements, err := spanParser.ParseStatements(cs.Statements)
			if err != nil {
				logger.Error("Error parsing span statements", zap.Error(err))
				errors = multierr.Append(errors, err)
				continue
			}

			s := createSampler(cs.SamplingConfig)
			if s != nil {
				err = s.Start()
				if err != nil {
					logger.Error("Error starting sampler", zap.Error(err))
					errors = multierr.Append(errors, err)
					continue
				}
			}

			transformations.spanTransforms[cs.RuleId] = &spanTransform{
				context:    cs.Context,
				conditions: conditions,
				statements: statements,
				sampler:    s,
				version:    cs.Version,
			}

		case "metric":
			conditions, err := metricParser.ParseConditions(cs.Conditions)
			if err != nil {
				logger.Error("Error parsing metric conditions", zap.Error(err))
				errors = multierr.Append(errors, err)
				continue
			}
			statements, err := metricParser.ParseStatements(cs.Statements)
			if err != nil {
				logger.Error("Error parsing metric statements", zap.Error(err))
				errors = multierr.Append(errors, err)
				continue
			}

			transformations.metricTransforms[cs.RuleId] = &metricTransform{
				context:    cs.Context,
				conditions: conditions,
				statements: statements,
				version:    cs.Version,
			}

		case "datapoint":
			conditions, err := dataPointParser.ParseConditions(cs.Conditions)
			if err != nil {
				logger.Error("Error parsing datapoint conditions", zap.Error(err))
				errors = multierr.Append(errors, err)
				continue
			}
			statements, err := dataPointParser.ParseStatements(cs.Statements)
			if err != nil {
				logger.Error("Error parsing datapoint statements", zap.Error(err))
				errors = multierr.Append(errors, err)
				continue
			}

			transformations.dataPointTransforms[cs.RuleId] = &dataPointTransform{
				context:    cs.Context,
				conditions: conditions,
				statements: statements,
				version:    cs.Version,
			}

		default:
			logger.Error("Unknown context: ", zap.String("context", string(cs.Context)))
		}
	}

	return transformations, errors
}

func evaluateTransform[V Versioned](tele *Telemetry, rules map[RuleID]V, eval func(*Telemetry, V, string, int)) {
	for ruleID, transform := range rules {
		eval(tele, transform, string(ruleID), transform.GetVersion())
	}
}

func setstokvs(sets ...attribute.Set) []attribute.KeyValue {
	kvs := []attribute.KeyValue{}
	for _, set := range sets {
		kvs = append(kvs, set.ToSlice()...)
	}
	return kvs
}

func (t *transformations) ExecuteResourceTransforms(logger *zap.Logger, incset attribute.Set, tele *Telemetry, transformCtx ottlresource.TransformContext) {
	kvs := setstokvs(incset)
	kvs = append(kvs, attribute.String("context", "resource"))
	evaluateTransform(tele, t.resourceTransforms, func(tele *Telemetry, resourceTransform *resourceTransform, ruleID string, version int) {
		startTime := time.Now()
		attrset := attribute.NewSet(append(kvs, attribute.String(translate.RuleId, ruleID), attribute.Int(translate.Version, version))...)
		allConditionsTrue := true
		for _, condition := range resourceTransform.conditions {
			conditionMet, err := condition.Eval(context.Background(), transformCtx)
			if err != nil {
				telemetry.CounterAdd(tele.ConditionsErrorCounter, 1, metric.WithAttributeSet(attrset))
				logger.Error("Error executing resource conditions", zap.Error(err))
			}
			allConditionsTrue = allConditionsTrue && conditionMet
		}
		telemetry.HistogramRecord(tele.ConditionsEvaluatedHistogram, time.Since(startTime).Nanoseconds(), metric.WithAttributeSet(attrset))
		telemetry.CounterAdd(tele.ConditionsEvaluatedCounter, 1, metric.WithAttributeSet(attrset))
		if !allConditionsTrue {
			return
		}

		startTime = time.Now()
		for _, statement := range resourceTransform.statements {
			_, _, err := statement.Execute(context.Background(), transformCtx)
			if err != nil {
				telemetry.CounterAdd(tele.StatementsErrorCounter, 1, metric.WithAttributeSet(attrset))
				logger.Error("Error executing resource transformation", zap.Error(err))
			}
		}
		telemetry.CounterAdd(tele.StatementsExecutedCounter, 1, metric.WithAttributeSet(attrset))
		telemetry.HistogramRecord(tele.StatementsExecutedHistogram, time.Since(startTime).Nanoseconds(), metric.WithAttributeSet(attrset))
	})
}

func (t *transformations) ExecuteScopeTransforms(logger *zap.Logger, incset attribute.Set, tele *Telemetry, transformCtx ottlscope.TransformContext) {
	kvs := setstokvs(incset)
	kvs = append(kvs, attribute.String("context", "scope"))
	evaluateTransform(tele, t.scopeTransforms, func(tele *Telemetry, scopeTransform *scopeTransform, ruleID string, version int) {
		startTime := time.Now()
		attrset := attribute.NewSet(append(kvs, attribute.String(translate.RuleId, ruleID), attribute.Int(translate.Version, version))...)
		allConditionsTrue := true
		for _, condition := range scopeTransform.conditions {
			conditionMet, err := condition.Eval(context.Background(), transformCtx)
			if err != nil {
				telemetry.CounterAdd(tele.ConditionsErrorCounter, 1, metric.WithAttributeSet(attrset))
				logger.Error("Error executing scope conditions", zap.Error(err))
			}
			allConditionsTrue = allConditionsTrue && conditionMet
		}
		telemetry.HistogramRecord(tele.ConditionsEvaluatedHistogram, time.Since(startTime).Nanoseconds(), metric.WithAttributeSet(attrset))
		telemetry.CounterAdd(tele.ConditionsEvaluatedCounter, 1, metric.WithAttributeSet(attrset))
		if !allConditionsTrue {
			return
		}

		startTime = time.Now()
		for _, statement := range scopeTransform.statements {
			_, _, err := statement.Execute(context.Background(), transformCtx)
			if err != nil {
				telemetry.CounterAdd(tele.StatementsErrorCounter, 1, metric.WithAttributeSet(attrset))
				logger.Error("Error executing scope transformation", zap.Error(err))
			}
		}
		telemetry.CounterAdd(tele.StatementsExecutedCounter, 1, metric.WithAttributeSet(attrset))
		telemetry.HistogramRecord(tele.StatementsExecutedHistogram, time.Since(startTime).Nanoseconds(), metric.WithAttributeSet(attrset))
	})
}

func (t *transformations) ExecuteLogTransforms(logger *zap.Logger, incset attribute.Set, tele *Telemetry, transformCtx ottllog.TransformContext) {
	kvs := setstokvs(incset)
	kvs = append(kvs, attribute.String("context", "log"))

	evaluateTransform(tele, t.logTransforms, func(tele *Telemetry, logTransform *logTransform, ruleID string, version int) {
		attrset := attribute.NewSet(append(kvs, attribute.String(translate.RuleId, ruleID), attribute.Int(translate.Version, version))...)
		startTime := time.Now()
		allConditionsTrue := true
		for _, condition := range logTransform.conditions {
			conditionMet, err := condition.Eval(context.Background(), transformCtx)
			if err != nil {
				telemetry.CounterAdd(tele.ConditionsErrorCounter, 1, metric.WithAttributeSet(attrset))
				logger.Error("Error executing log conditions", zap.Error(err))
			}
			allConditionsTrue = allConditionsTrue && conditionMet
		}
		telemetry.HistogramRecord(tele.ConditionsEvaluatedHistogram, time.Since(startTime).Nanoseconds(), metric.WithAttributeSet(attrset))
		telemetry.CounterAdd(tele.ConditionsEvaluatedCounter, 1, metric.WithAttributeSet(attrset))
		if !allConditionsTrue {
			return
		}

		var shouldAllow = true
		if logTransform.sampler != nil {
			serviceName := GetServiceName(transformCtx.GetResource())
			fingerprint, exists := transformCtx.GetLogRecord().Attributes().Get(translate.CardinalFieldFingerprint)
			if !exists {
				return
			}
			key := fmt.Sprintf("%s:%s", serviceName, fingerprint.AsString())
			sampleRate := logTransform.sampler.GetSampleRate(key)
			shouldAllow = shouldFilter(sampleRate, rand.Float64())
		}
		if !shouldAllow {
			telemetry.CounterAdd(tele.RateLimitedCounter, 1, metric.WithAttributeSet(attrset))
			return
		}

		startTime = time.Now()
		for _, statement := range logTransform.statements {
			_, _, err := statement.Execute(context.Background(), transformCtx)
			if err != nil {
				telemetry.CounterAdd(tele.StatementsErrorCounter, 1, metric.WithAttributeSet(attrset))
				logger.Error("Error executing log transformation", zap.Error(err))
			}
		}
		telemetry.CounterAdd(tele.StatementsExecutedCounter, 1, metric.WithAttributeSet(attrset))
		telemetry.HistogramRecord(tele.StatementsExecutedHistogram, time.Since(startTime).Nanoseconds(), metric.WithAttributeSet(attrset))
	})
}

func shouldFilter(rate int, randval float64) bool {
	switch rate {
	case 0:
		return true
	case 1:
		return false
	default:
		return randval > 1/float64(rate)
	}
}

func (t *transformations) ExecuteSpanTransforms(logger *zap.Logger, incset attribute.Set, tele *Telemetry, transformCtx ottlspan.TransformContext) {
	kvs := setstokvs(incset)
	kvs = append(kvs, attribute.String("context", "span"))
	evaluateTransform(tele, t.spanTransforms, func(tele *Telemetry, spanTransform *spanTransform, ruleID string, version int) {
		attrset := attribute.NewSet(append(kvs, attribute.String(translate.RuleId, ruleID), attribute.Int(translate.Version, version))...)
		startTime := time.Now()

		allConditionsTrue := true
		for _, condition := range spanTransform.conditions {
			conditionMet, err := condition.Eval(context.Background(), transformCtx)
			if err != nil {
				telemetry.CounterAdd(tele.ConditionsErrorCounter, 1, metric.WithAttributeSet(attrset))
				logger.Error("Error executing span conditions", zap.Error(err))
			}
			allConditionsTrue = allConditionsTrue && conditionMet
		}
		telemetry.HistogramRecord(tele.ConditionsEvaluatedHistogram, time.Since(startTime).Nanoseconds(), metric.WithAttributeSet(attrset))
		telemetry.CounterAdd(tele.ConditionsEvaluatedCounter, 1, metric.WithAttributeSet(attrset))
		if !allConditionsTrue {
			return
		}

		var shouldAllow = true
		if spanTransform.sampler != nil {
			randval := rand.Float64()
			serviceName := GetServiceName(transformCtx.GetResource())
			fingerprint, exists := transformCtx.GetSpan().Attributes().Get(translate.CardinalFieldFingerprint)
			if !exists {
				return
			}
			key := fmt.Sprintf("%s:%s", serviceName, fingerprint.AsString())
			sampleRate := spanTransform.sampler.GetSampleRate(key)
			shouldAllow = shouldFilter(sampleRate, randval)
		}
		if !shouldAllow {
			telemetry.CounterAdd(tele.RateLimitedCounter, 1, metric.WithAttributeSet(attrset))
			return
		}

		startTime = time.Now()
		for _, statement := range spanTransform.statements {
			_, _, err := statement.Execute(context.Background(), transformCtx)
			if err != nil {
				telemetry.CounterAdd(tele.StatementsErrorCounter, 1, metric.WithAttributeSet(attrset))
				logger.Error("Error executing span transformation", zap.Error(err))
			}
		}
		telemetry.CounterAdd(tele.StatementsExecutedCounter, 1, metric.WithAttributeSet(attrset))
		telemetry.HistogramRecord(tele.StatementsExecutedHistogram, time.Since(startTime).Nanoseconds(), metric.WithAttributeSet(attrset))
	})
}

func (t *transformations) ExecuteMetricTransforms(logger *zap.Logger, incset attribute.Set, tele *Telemetry, transformCtx ottlmetric.TransformContext) {
	kvs := setstokvs(incset)
	kvs = append(kvs, attribute.String("context", "metric"))
	evaluateTransform(tele, t.metricTransforms, func(tele *Telemetry, metricTransform *metricTransform, ruleID string, version int) {
		attrset := attribute.NewSet(append(kvs, attribute.String(translate.RuleId, ruleID), attribute.Int(translate.Version, version))...)
		startTime := time.Now()

		allConditionsTrue := true
		for _, condition := range metricTransform.conditions {
			conditionMet, err := condition.Eval(context.Background(), transformCtx)
			if err != nil {
				telemetry.CounterAdd(tele.ConditionsErrorCounter, 1, metric.WithAttributeSet(attrset))
				logger.Error("Error executing metric conditions", zap.Error(err))
			}
			allConditionsTrue = allConditionsTrue && conditionMet
		}
		telemetry.HistogramRecord(tele.ConditionsEvaluatedHistogram, time.Since(startTime).Nanoseconds(), metric.WithAttributeSet(attrset))
		telemetry.CounterAdd(tele.ConditionsEvaluatedCounter, 1, metric.WithAttributeSet(attrset))
		if !allConditionsTrue {
			return
		}
		startTime = time.Now()
		for _, statement := range metricTransform.statements {
			_, _, err := statement.Execute(context.Background(), transformCtx)
			if err != nil {
				telemetry.CounterAdd(tele.StatementsErrorCounter, 1, metric.WithAttributeSet(attrset))
				logger.Error("Error executing metric transformation", zap.Error(err))
			}
		}
		telemetry.CounterAdd(tele.StatementsExecutedCounter, 1, metric.WithAttributeSet(attrset))
		telemetry.HistogramRecord(tele.StatementsExecutedHistogram, time.Since(startTime).Nanoseconds(), metric.WithAttributeSet(attrset))
	})
}

func (t *transformations) ExecuteDatapointTransforms(logger *zap.Logger, incset attribute.Set, tele *Telemetry, transformCtx ottldatapoint.TransformContext) {
	kvs := setstokvs(incset)
	kvs = append(kvs, attribute.String("context", "datapoint"))
	evaluateTransform(tele, t.dataPointTransforms, func(tele *Telemetry, dataPointTransform *dataPointTransform, ruleID string, version int) {
		attrset := attribute.NewSet(append(kvs, attribute.String(translate.RuleId, ruleID), attribute.Int(translate.Version, version))...)
		startTime := time.Now()

		allConditionsTrue := true
		for _, condition := range dataPointTransform.conditions {
			conditionMet, err := condition.Eval(context.Background(), transformCtx)
			if err != nil {
				telemetry.CounterAdd(tele.ConditionsErrorCounter, 1, metric.WithAttributeSet(attrset))
				logger.Error("Error executing span conditions", zap.Error(err))
			}
			allConditionsTrue = allConditionsTrue && conditionMet
		}
		telemetry.HistogramRecord(tele.ConditionsEvaluatedHistogram, time.Since(startTime).Nanoseconds(), metric.WithAttributeSet(attrset))
		telemetry.CounterAdd(tele.ConditionsEvaluatedCounter, 1, metric.WithAttributeSet(attrset))
		if !allConditionsTrue {
			return
		}

		startTime = time.Now()
		for _, statement := range dataPointTransform.statements {
			_, _, err := statement.Execute(context.Background(), transformCtx)
			if err != nil {
				telemetry.CounterAdd(tele.StatementsErrorCounter, 1, metric.WithAttributeSet(attrset))
				logger.Error("Error executing datapoint transformation", zap.Error(err))
			}
		}
		telemetry.CounterAdd(tele.StatementsExecutedCounter, 1, metric.WithAttributeSet(attrset))
		telemetry.HistogramRecord(tele.StatementsExecutedHistogram, time.Since(startTime).Nanoseconds(), metric.WithAttributeSet(attrset))
	})
}
