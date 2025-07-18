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

package dateutils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	relativeDateRegex = regexp.MustCompile(`^([a-z]+)([-+])(.+)$`)
	unixDateRegex     = regexp.MustCompile(`^\d{10,}$`)
	namedDateRegex    = regexp.MustCompile(`^(now|epoch|e)$`)
)

// ToStartEnd converts optional start/end strings into milliseconds since epoch.
func ToStartEnd(s, e string) (int64, int64, error) {
	if s == "" {
		s = "e-1h"
	}
	if e == "" {
		e = "now"
	}
	startTime, endTime, err := TimeRange(s, e, time.UTC)
	if err != nil {
		return 0, 0, err
	}
	return startTime.UnixMilli(), endTime.UnixMilli(), nil
}

// TimeRange returns the start and end time.Time from input strings with timezone context.
func TimeRange(s, e string, tz *time.Location) (time.Time, time.Time, error) {
	var start, end time.Time
	var err error

	if IsRelativeDate(s, true) || s == "e" {
		if IsRelativeDate(e, true) {
			return time.Time{}, time.Time{}, errors.New("start and end are both relative")
		}
		end, err = ParseDate(time.Time{}, e, tz)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		start, err = ParseDate(end, s, tz)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	} else {
		start, err = ParseDate(time.Time{}, s, tz)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		end, err = ParseDate(start, e, tz)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	if !IsBeforeOrEqual(start, end) {
		return time.Time{}, time.Time{}, errors.New("end time is before start time")
	}
	return start, end, nil
}

// ParseDate resolves time string to a time.Time, given a reference and timezone.
func ParseDate(ref time.Time, s string, tz *time.Location) (time.Time, error) {
	s = strings.ToLower(strings.TrimSpace(s))

	if strings.Contains(s, "t") || strings.HasSuffix(s, "z") {
		s = strings.Replace(s, "t", "T", 1)
		if strings.HasSuffix(s, "z") {
			s = strings.TrimSuffix(s, "z") + "Z"
		}
	}
	// Relative format like e-1h
	if matches := relativeDateRegex.FindStringSubmatch(s); len(matches) == 4 {
		refName, op, durationStr := matches[1], matches[2], matches[3]
		base, err := parseRefVar(ref, refName, tz)
		if err != nil {
			return time.Time{}, err
		}
		dur, err := parseDuration(durationStr)
		if err != nil {
			return time.Time{}, err
		}
		if op == "-" {
			return base.Add(-dur), nil
		}
		return base.Add(dur), nil
	}

	// Named reference
	if namedDateRegex.MatchString(s) {
		return parseRefVar(ref, s, tz)
	}

	// Unix timestamp
	if unixDateRegex.MatchString(s) {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid unix timestamp: %w", err)
		}
		if v <= int64(2147483647) {
			v *= 1000
		}
		return time.UnixMilli(v).In(tz), nil
	}

	// ISO-8601 or fallback to RFC3339
	t, err := time.ParseInLocation(time.RFC3339, s, tz)
	if err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date: %s", s)
}

// IsRelativeDate checks if a string is in relative format and allowed by config.
func IsRelativeDate(s string, customRef bool) bool {
	matches := relativeDateRegex.FindStringSubmatch(s)
	if len(matches) == 4 {
		ref := matches[1]
		return !customRef || (ref != "now" && ref != "epoch")
	}
	return false
}

// IsBeforeOrEqual returns true if s <= e
func IsBeforeOrEqual(s, e time.Time) bool {
	return s.Before(e) || s.Equal(e)
}

// parseRefVar maps known reference strings to base time instants.
func parseRefVar(ref time.Time, s string, tz *time.Location) (time.Time, error) {
	switch s {
	case "now":
		return time.Now().In(tz), nil
	case "epoch":
		return time.Unix(0, 0).In(tz), nil
	case "e":
		if ref.IsZero() {
			return time.Time{}, fmt.Errorf("reference time 'e' not available")
		}
		return ref.In(tz), nil
	default:
		return time.Time{}, fmt.Errorf("unrecognized ref var: %s", s)
	}
}

// parseDuration parses a simplified duration like "1h", "2d", "15m", "3w".
func parseDuration(s string) (time.Duration, error) {
	// Support "d" (days) and "w" (weeks)
	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, err
		}
		return time.Hour * 24 * time.Duration(n), nil
	}
	if strings.HasSuffix(s, "w") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "w"))
		if err != nil {
			return 0, err
		}
		return time.Hour * 24 * 7 * time.Duration(n), nil
	}
	// Fall back to Go's native duration parsing
	return time.ParseDuration(s)
}
