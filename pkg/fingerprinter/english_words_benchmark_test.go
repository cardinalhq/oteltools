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

package fingerprinter

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/cardinalhq/oteltools/pkg/fingerprinter/wordlist"
)

// Test words that should exist in the dictionary (common English words)
var existingWords = []string{
	"the", "and", "for", "are", "but", "not", "you", "all", "can", "had",
	"her", "was", "one", "our", "out", "day", "get", "has", "him", "his",
	"how", "man", "new", "now", "old", "see", "two", "way", "who", "boy",
	"did", "its", "let", "put", "say", "she", "too", "use", "what", "when",
	"where", "which", "with", "have", "this", "will", "your", "from", "they",
	"know", "want", "been", "good", "much", "some", "time", "very", "come",
	"could", "would", "there", "people", "first", "water", "after", "back",
	"little", "only", "right", "think", "also", "around", "another", "came",
	"work", "three", "word", "long", "great", "small", "every", "start",
	"place", "made", "live", "where", "after", "back", "little", "only",
	"round", "right", "think", "also", "just", "should", "before", "turn",
	"sentence", "different", "small", "large", "spell", "again", "animal",
	"house", "point", "page", "letter", "mother", "answer", "found", "study",
	"still", "learn", "should", "america", "world", "high", "every", "near",
	"add", "food", "between", "own", "below", "country", "plant", "last",
	"school", "father", "keep", "tree", "never", "start", "city", "earth",
}

// Test words that should NOT exist in the dictionary (almost-words that share prefixes with real words)
// These exercise trie traversal by following valid paths before diverging
// Using repeated words with 'x' separator - definitely won't exist but exercise full trie paths
var nonExistingWords = []string{
	// Based on common words repeated with 'x' separator
	"thexthe", "andxand", "forxfor", "butxbut", "notxnot",
	"youxyou", "allxall", "canxcan", "hadxhad", "herxher",
	"wasxwas", "onexone", "ourxour", "outxout", "dayxday",
	"getxget", "hasxhas", "himxhim", "hisxhis", "howxhow",
	"doggoxdoggo", "newxnew", "nowxnow", "oldxold", "sexsex",
	"twoxtwo", "wayxway", "whoxwho", "boyxboy", "didxdid",
	// Longer words repeated with 'x' separator for deeper trie traversal
	"waterxwater", "peoplexpeople", "housexhouse", "schoolxschool", "workxwork",
	"thinkxthink", "worldxworld", "rightxright", "smallxsmall", "greatxgreat",
	"placexplace", "countryxcountry", "fatherxfather", "motherxmother", "treextree",
	// Mixed case and variations
	"timextime", "livexlive", "studyxstudy", "learnxlearn", "startxstart",
	"cityxcity", "earthxearth", "plantxplant", "animalxanimal", "foodxfood",
	// Even longer compound words for maximum trie depth
	"sentencexsentence", "differentxdifferent", "americaxamerica", "schoolxschool",
	"betweenxbetween", "belowxbelow", "neverxnever", "againxagain",
	// Some with shorter prefixes that diverge early
	"thxth", "anxan", "foxfo", "waxwa", "pexpe",
	"hoxho", "scxsc", "woxwo", "rixri", "smxsm",
}

func BenchmarkWordlistLookupExisting(b *testing.B) {
	// Create a random source
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Pick a random existing word
		word := existingWords[rng.Intn(len(existingWords))]
		_, exists := wordlist.EnglishWords[strings.ToLower(word)]
		if !exists {
			b.Errorf("Expected word %s to exist in dictionary", word)
		}
	}
}

func BenchmarkWordlistLookupNonExisting(b *testing.B) {
	// Create the wordlist once
	// Create a random source
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Pick a random non-existing word
		word := nonExistingWords[rng.Intn(len(nonExistingWords))]
		_, exists := wordlist.EnglishWords[strings.ToLower(word)]
		if exists {
			b.Errorf("Expected word %s to NOT exist in dictionary", word)
		}
	}
}

func BenchmarkWordlistLookupMixed(b *testing.B) {
	// Create the wordlist once
	// Create a random source
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var word string
		// 50% chance of existing word, 50% chance of non-existing word
		if rng.Float32() < 0.5 {
			word = existingWords[rng.Intn(len(existingWords))]
		} else {
			word = nonExistingWords[rng.Intn(len(nonExistingWords))]
		}
		_ = wordlist.EnglishWords[strings.ToLower(word)]
	}
}

func BenchmarkFingerprinterCreation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		fp := NewFingerprinter()
		_ = fp
	}
}

func BenchmarkFingerprinterIsWord(b *testing.B) {
	fp := NewFingerprinter()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var word string
		// 50% chance of existing word, 50% chance of non-existing word
		if rng.Float32() < 0.5 {
			word = existingWords[rng.Intn(len(existingWords))]
		} else {
			word = nonExistingWords[rng.Intn(len(nonExistingWords))]
		}
		_ = fp.IsWord(word)
	}
}

func BenchmarkMemoryUsageMultipleFingerprinters(b *testing.B) {
	b.ReportAllocs()

	// Create multiple fingerprinters to measure memory overhead
	fingerprinters := make([]*fingerprinterImpl, b.N)
	for i := 0; i < b.N; i++ {
		fingerprinters[i] = NewFingerprinter()
	}

	// Prevent optimization away
	_ = fingerprinters
}
