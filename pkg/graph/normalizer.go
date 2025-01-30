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

package graph

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"regexp"
)

func toMessagingEntities(messagingAttributes pcommon.Map) map[string]*ResourceEntity {
	messagingSystemKey := string(semconv.MessagingSystemKey)
	messagingSystem, messagingSystemFound := messagingAttributes.Get(messagingSystemKey)

	if messagingSystemFound {
		entities := make(map[string]*ResourceEntity)
		consumerGroupKey := string(semconv.MessagingConsumerGroupNameKey)
		messagingConsumerGroup, messagingConsumerGroupFound := messagingAttributes.Get(consumerGroupKey)
		messagingDestination, messagingDestinationFound := messagingAttributes.Get(string(semconv.MessagingDestinationNameKey))
		serverAddress := getValue(messagingAttributes, string(semconv.ServerAddressKey), string(semconv.NetworkPeerAddressKey), NetPeerName, string(semconv.RPCServiceKey))

		if messagingConsumerGroupFound {
			consumerGroupEntity := &ResourceEntity{
				AttributeName: consumerGroupKey,
				Name:          messagingConsumerGroup.AsString(),
				Type:          MessagingConsumer,
				Attributes:    map[string]string{},
			}
			if serverAddress != "" {
				consumerGroupEntity.Attributes[string(semconv.ServerAddressKey)] = serverAddress
				consumerGroupEntity.Attributes[string(semconv.MessagingSystemKey)] = messagingSystem.AsString()
			}
			entities[toEntityId(consumerGroupEntity.Name, consumerGroupEntity.Type)] = consumerGroupEntity
		}
		if messagingDestinationFound {
			destinationEntity := &ResourceEntity{
				AttributeName: string(semconv.MessagingDestinationNameKey),
				Name:          messagingDestination.AsString(),
				Type:          MessagingProducer,
				Attributes:    map[string]string{},
			}
			if serverAddress != "" {
				destinationEntity.Attributes[string(semconv.ServerAddressKey)] = serverAddress
				destinationEntity.Attributes[string(semconv.MessagingSystemKey)] = messagingSystem.AsString()
			}
			entities[toEntityId(destinationEntity.Name, destinationEntity.Type)] = destinationEntity
		}
		return entities
	}
	return nil
}

func toDBEntities(dbAttributes pcommon.Map) map[string]*ResourceEntity {
	dbSystemKey := string(semconv.DBSystemKey)
	dbSystem, dbSystemFound := dbAttributes.Get(dbSystemKey)
	if dbSystemFound {
		entities := make(map[string]*ResourceEntity)

		dbSystemValue := dbSystem.AsString()

		// Get DB instance name from the attributes
		dbInstance := getValue(dbAttributes, string(semconv.ServerAddressKey),
			string(semconv.NetworkPeerAddressKey),
			NetPeerName,
			string(semconv.RPCServiceKey))

		instanceEntityId := ""
		if dbInstance != "" {
			instanceEntityId = toEntityId(dbInstance, dbSystemValue)
			dbAttributes.PutStr(string(semconv.NetworkPeerAddressKey), dbInstance)
			entities[instanceEntityId] = &ResourceEntity{
				AttributeName: string(semconv.DBSystemKey),
				Name:          dbInstance,
				Type:          dbSystemValue,
			}
		} else {
			instanceEntityId = toEntityId(dbSystemValue, dbSystemKey)
			entities[instanceEntityId] = &ResourceEntity{
				AttributeName: dbSystemKey,
				Name:          dbSystemValue,
				Type:          dbSystemKey,
			}
		}

		dbNamespaceKey := string(semconv.DBNamespaceKey)
		dbNamespace, dbNamespaceFound := dbAttributes.Get(dbNamespaceKey)
		namespaceEntityId := ""
		if dbNamespaceFound {
			dbNamespaceValue := dbNamespace.AsString()
			namespaceEntityId = toEntityId(dbNamespaceValue, dbNamespaceKey)
			entities[namespaceEntityId] = &ResourceEntity{
				AttributeName: dbNamespaceKey,
				Name:          dbNamespaceValue,
				Type:          dbNamespaceKey,
			}
			entities[instanceEntityId].AddEdge(dbNamespaceValue, dbNamespaceKey, HasNamespace)
			entities[namespaceEntityId].AddEdge(dbInstance, dbSystemValue, IsDatabaseHostedOn)
		}

		dbCollection := getValue(dbAttributes, string(semconv.DBCollectionNameKey), string(semconv.AWSDynamoDBTableNamesKey))
		if dbCollection != "" {
			dbAttributes.PutStr(string(semconv.DBCollectionNameKey), dbCollection)
			collectionEntityId := toEntityId(dbCollection, string(semconv.DBCollectionNameKey))
			entities[collectionEntityId] = &ResourceEntity{
				AttributeName: string(semconv.DBCollectionNameKey),
				Name:          dbCollection,
				Type:          string(semconv.DBCollectionNameKey),
			}
			if dbNamespaceFound {
				entities[namespaceEntityId].AddEdge(dbCollection, string(semconv.DBCollectionNameKey), HasCollection)
				entities[collectionEntityId].AddEdge(dbNamespace.AsString(), dbNamespaceKey, IsCollectionHostedOn)
			}
		}

		dbQuery := getValue(dbAttributes, string(semconv.DBQueryTextKey), DBStatement, DBQuerySummary)
		if dbQuery != "" {
			dbAttributes.PutStr(string(semconv.DBQueryTextKey), normalizeSQL(dbQuery))
		}
		return entities
	}
	return nil
}

func normalizeSQL(query string) string {
	// Regex to match:
	// - Numbers (\b\d+\b)
	// - Single-quoted strings ('[^']*')
	// - Double-quoted strings ("[^"]*")
	pattern := `\b\d+\.\d+\b|\b\d+\b|'[^']*'|"[^"]*"`
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(query, "?")
}

func getValue(attributes pcommon.Map, fallbackChain ...string) string {
	for _, key := range fallbackChain {
		if value, found := attributes.Get(key); found {
			return value.AsString()
		}
	}
	return ""
}
