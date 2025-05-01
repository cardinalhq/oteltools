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

package functions

import (
	"context"
	"fmt"
	"net/netip"
	"os"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/oschwald/maxminddb-golang/v2"
)

// IpLocationArguments holds the target IP
type IpLocationArguments[K any] struct {
	Target ottl.StringGetter[K] // Target is the IP address to locate
}

// NewIpLocationFactory Create the factory for the iplocation function
func NewIpLocationFactory[K any]() ottl.Factory[K] {
	// This is done here to avoid opening the database for every function call.
	// Any database errors will be returned when the factory is created.
	dbPath := os.Getenv("GEOIP_DB_PATH")
	if dbPath == "" {
		dbPath = "/app/geoip/GeoOpen-Country.mmdb"
	}
	db, dbopenError := maxminddb.Open(dbPath)

	return ottl.NewFactory("IpLocation", &IpLocationArguments[K]{}, func(ctx ottl.FunctionContext, args ottl.Arguments) (ottl.ExprFunc[K], error) {
		if dbopenError != nil {
			return nil, fmt.Errorf("failed to open database: %v", dbopenError)
		}
		ipArgs, ok := args.(*IpLocationArguments[K])
		if !ok {
			return nil, fmt.Errorf("iplocation args must be of type *IpLocationArguments[K]")
		}

		return createIpLocationFunction(db, ipArgs.Target)
	})
}

func createIpLocationFunction[K any](db *maxminddb.Reader, ipGetter ottl.StringGetter[K]) (ottl.ExprFunc[K], error) {

	return iplocation(db, ipGetter), nil
}

type IPLocationLookup struct {
	Country struct {
		ISOCode string            `maxminddb:"iso_code"`
		Names   map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Location struct {
		Latitude  float64 `maxminddb:"latitude"`
		Longitude float64 `maxminddb:"longitude"`
	} `maxminddb:"location"`
	Postal struct {
		Code string `maxminddb:"code"`
	} `maxminddb:"postal"`
}

// iplocation function performs the geolocation lookup
func iplocation[K any](db *maxminddb.Reader, ipGetter ottl.StringGetter[K]) ottl.ExprFunc[K] {
	return func(ctx context.Context, tCtx K) (any, error) {
		ip, err := ipGetter.Get(ctx, tCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to get IP address: %v", err)
		}

		parsedIP, err := netip.ParseAddr(ip)
		if err != nil {
			return nil, fmt.Errorf("failed to parse IP address: %v", err)
		}

		var record IPLocationLookup
		err = db.Lookup(parsedIP).Decode(&record)

		if err != nil {
			return nil, fmt.Errorf("error looking up IP: %v", err)
		}

		city := record.City.Names["en"]
		if city == "" {
			city = "Unknown"
		}

		country := record.Country.Names["en"]
		if country == "" {
			country = "Unknown"
		}

		location := map[string]any{
			"city":        city,
			"country":     country,
			"country_iso": record.Country.ISOCode,
			"zip_code":    record.Postal.Code,
			"latitude":    record.Location.Latitude,
			"longitude":   record.Location.Longitude,
		}

		return location, nil
	}
}
