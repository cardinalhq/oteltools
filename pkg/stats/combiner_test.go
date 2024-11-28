// Copyright 2024 CardinalHQ, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stats

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

func TestStatsCombiner_RecordAndFlush_SingleItem(t *testing.T) {
	db, err := bbolt.Open("test_single_item.db", 0666, nil)
	assert.NoError(t, err)
	defer func() {
		db.Close()
		os.Remove("test_single_item.db")
	}()

	serializer := func(obj *MockStatsObject) ([]byte, error) {
		marshal, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		return marshal, nil
	}
	deserializer := func(data []byte) (*MockStatsObject, error) {
		var obj MockStatsObject
		err := json.Unmarshal(data, &obj)
		return &obj, err
	}

	now := time.Now()
	combiner := NewStatsCombiner[*MockStatsObject](
		db, "stats_test", now, 10*time.Second, serializer, deserializer,
	)

	item := &MockStatsObject{Id: 1, Count: 1, Size: 10}
	flushed, err := combiner.Record(now, item, "key1", 1, 10)
	assert.NoError(t, err)
	assert.Nil(t, flushed)
	flushed1, err1 := combiner.Record(now, item, "key1", 1, 10)
	assert.NoError(t, err1)
	assert.Nil(t, flushed1)

	flushed, err = combiner.Record(now.Add(15*time.Second), item, "key1", 1, 10)
	assert.NoError(t, err)
	assert.NotNil(t, flushed)

	assert.Len(t, flushed, 1)
	assert.Len(t, flushed[1], 1)
	assert.Equal(t, 2, flushed[1][0].Count)
	assert.Equal(t, int64(20), flushed[1][0].Size)
}
