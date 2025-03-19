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
	"errors"
	"fmt"

	"github.com/cardinalhq/oteltools/pkg/ottl/accumulator"
)

type AggregationType int

const (
	AggregationTypeSum AggregationType = iota
	AggregationTypeAvg
	AggregationTypeMin
	AggregationTypeMax
)

func (a AggregationType) String() string {
	switch a {
	case AggregationTypeSum:
		return "sum"
	case AggregationTypeAvg:
		return "avg"
	case AggregationTypeMin:
		return "min"
	case AggregationTypeMax:
		return "max"
	default:
		return "unknown"
	}
}

var stringToAggregationType = map[string]AggregationType{
	"sum": AggregationTypeSum,
	"avg": AggregationTypeAvg,
	"min": AggregationTypeMin,
	"max": AggregationTypeMax,
}

func ParseAggregationType(s string) (AggregationType, error) {
	if aggType, exists := stringToAggregationType[s]; exists {
		return aggType, nil
	}
	return 0, errors.New("invalid aggregation type")
}

type Aggregation interface {
	Add(name string, values []float64) error
	Value() []float64
	Count() uint64
	Tags() map[string]string
	Buckets() []float64
}

type AggregationImpl[T int64 | float64] struct {
	ty          AggregationType
	name        string
	tags        map[string]string
	accumulator accumulator.Accumulator[T]
}

func NewAggregationImpl[T int64 | float64](name string, buckets []T, aggregationType AggregationType, tags map[string]string) *AggregationImpl[T] {
	return &AggregationImpl[T]{
		name:        name,
		ty:          aggregationType,
		tags:        tags,
		accumulator: accumulator.NewAccumulatorImlp(buckets),
	}
}

func (a *AggregationImpl[T]) Add(name string, values []T) error {
	if a.name != name {
		return fmt.Errorf("aggregation name mismatch: %s != %s", a.name, name)
	}
	return a.accumulator.Add(values)
}

func (a *AggregationImpl[T]) Value() []T {
	switch a.ty {
	case AggregationTypeAvg:
		return a.accumulator.Avg()
	case AggregationTypeMin:
		return a.accumulator.Min()
	case AggregationTypeMax:
		return a.accumulator.Max()
	default:
		return a.accumulator.Sum()
	}
}

func (a *AggregationImpl[T]) Count() uint64 {
	return a.accumulator.Count()
}

func (a *AggregationImpl[T]) Tags() map[string]string {
	return a.tags
}

func (a *AggregationImpl[T]) Buckets() []T {
	return a.accumulator.Buckets()
}

func (a *AggregationImpl[T]) AggregationType() AggregationType {
	return a.ty
}

func (a *AggregationImpl[T]) Name() string {
	return a.name
}
