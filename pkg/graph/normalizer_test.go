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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

func TestToDBEntities(t *testing.T) {
	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.DBSystemKey), "mysql")
	attributes.PutStr(NetPeerName, "ams-aiops-poc.cp534rias9ed.us-east-1.rds.amazonaws.com")
	attributes.PutStr(string(semconv.DBNamespaceKey), "ams_aiops_ams")
	attributes.PutStr(string(semconv.DBCollectionNameKey), "alerts")

	entities := toDBEntities(attributes)

	expectedEntities := map[string]*ResourceEntity{
		toEntityId("ams-aiops-poc.cp534rias9ed.us-east-1.rds.amazonaws.com", "mysql"): {
			AttributeName: string(semconv.DBSystemKey),
			Name:          "ams-aiops-poc.cp534rias9ed.us-east-1.rds.amazonaws.com",
			Type:          "mysql",
		},
		toEntityId("ams_aiops_ams", string(semconv.DBNamespaceKey)): {
			AttributeName: string(semconv.DBNamespaceKey),
			Name:          "ams_aiops_ams",
			Type:          string(semconv.DBNamespaceKey),
		},
		toEntityId("alerts", string(semconv.DBCollectionNameKey)): {
			AttributeName: string(semconv.DBCollectionNameKey),
			Name:          "alerts",
			Type:          string(semconv.DBCollectionNameKey),
		},
	}

	for entityId, expectedEntity := range expectedEntities {
		actualEntity, exists := entities[entityId]
		assert.True(t, exists, "Expected entity %s not found", entityId)
		assert.Equal(t, expectedEntity.AttributeName, actualEntity.AttributeName, "Mismatch in entity attribute name for %s", entityId)
		assert.Equal(t, expectedEntity.Name, actualEntity.Name, "Mismatch in entity name for %s", entityId)
		assert.Equal(t, expectedEntity.Type, actualEntity.Type, "Mismatch in entity type for %s", entityId)
	}
}

func TestDatabaseEntityRelationships(t *testing.T) {
	ec := NewResourceEntityCache()

	resourceAttributes := pcommon.NewMap()
	resourceAttributes.PutStr(string(semconv.ServiceNameKey), "orders-service")
	resourceAttributes.PutStr(string(semconv.K8SNamespaceNameKey), "orders-namespace")

	dbAttributes := pcommon.NewMap()
	dbAttributes.PutStr(string(semconv.DBSystemKey), "mysql")                                             // Database system
	dbAttributes.PutStr(string(semconv.NetworkPeerAddressKey), "db-instance-1")                           // Instance of MySQL
	dbAttributes.PutStr(string(semconv.DBNamespaceKey), "orders-db")                                      // Logical database name
	dbAttributes.PutStr(string(semconv.DBCollectionNameKey), "transactions")                              // Table/collection
	dbAttributes.PutStr(string(semconv.DBQueryTextKey), "SELECT * FROM transactions WHERE amount > 100;") // Query to be normalized

	globalEntityMap := ec.ProvisionResourceAttributes(resourceAttributes)
	ec.ProvisionRecordAttributes(globalEntityMap, dbAttributes)

	entities := ec._allEntities()

	expectedEntities := map[string]string{
		"orders-service": "service",
		"db-instance-1":  "mysql",
		"orders-db":      "db.namespace",
		"transactions":   "db.collection.name",
	}

	for name, entityType := range expectedEntities {
		entityID := toEntityId(name, entityType)
		entity, exists := entities[entityID]
		assert.True(t, exists, "Expected entity %s not found", entityID)
		assert.Equal(t, entityType, entity.Type, "Incorrect type for entity %s", entityID)
	}

	assert.Equal(t, UsesDatabase, entities[toEntityId("orders-service", "service")].Edges[toEntityId("db-instance-1", "mysql")].Relationship)
	assert.Equal(t, HasNamespace, entities[toEntityId("db-instance-1", "mysql")].Edges[toEntityId("orders-db", "db.namespace")].Relationship)
	assert.Equal(t, IsCollectionHostedOn, entities[toEntityId("transactions", "db.collection.name")].Edges[toEntityId("orders-db", "db.namespace")].Relationship)
}

func TestMessagingEntityRelationships(t *testing.T) {
	ec := NewResourceEntityCache()

	// Define resource attributes for a service
	resourceAttributes := pcommon.NewMap()
	resourceAttributes.PutStr(string(semconv.ServiceNameKey), "payment-service")
	resourceAttributes.PutStr(string(semconv.K8SNamespaceNameKey), "payments-namespace")

	// Define messaging attributes
	messagingAttributes := pcommon.NewMap()
	messagingAttributes.PutStr(string(semconv.MessagingSystemKey), "kafka")
	messagingAttributes.PutStr(string(semconv.MessagingDestinationNameKey), "payments-topic")
	messagingAttributes.PutStr(string(semconv.MessagingConsumerGroupNameKey), "payment-group")
	messagingAttributes.PutStr(string(semconv.NetworkPeerAddressKey), "kafka-broker-1")

	globalEntityMap := ec.ProvisionResourceAttributes(resourceAttributes)
	ec.ProvisionRecordAttributes(globalEntityMap, messagingAttributes)

	entities := ec._allEntities()

	expectedEntities := map[string]string{
		"payment-service": "service",
		"payments-topic":  "kafka",
		"payment-group":   "kafka",
	}

	for name, entityType := range expectedEntities {
		entityID := toEntityId(name, entityType)
		entity, exists := entities[entityID]
		assert.True(t, exists, "Expected entity %s not found", entityID)
		assert.Equal(t, entityType, entity.Type, "Incorrect type for entity %s", entityID)
	}

	assert.Equal(t, ProducesTo, entities[toEntityId("payment-service", "service")].Edges[toEntityId("payments-topic", "kafka")].Relationship)
	assert.Equal(t, ConsumesFrom, entities[toEntityId("payment-service", "service")].Edges[toEntityId("payment-group", "kafka")].Relationship)

}
