# GraphDB Design

## Features

### Overall Features

* Implementation detail:  Every object listed in this document is scoped to `organization_id` and most of the backing SQL tables are partitioned on that key.  Every query for any data will always include `organization_id`, `valid_from`, and `valid_to` (as either in the 'future' or NULL.)

* **Entities** represent nodes within the graph and contain attributes. Entities can be:
  * **Real**, such as Kubernetes Pods.
  * **Virtual**, used primarily for grouping, such as a Kubernetes Cluster.

* **Edges** represent directed relationships between entities, moving from a source entity to a target entity. Edges have relationship types that define their meaning, such as "manages," "uses," or "is used by."

* When entities are created, initial known edges and reverse links (if needed) are established.
  * **Direct edges**: Explicitly known relationships (e.g., a Pod's owner reference to its StatefulSet) are created immediately upon entity creation.
  * **Computed edges**: Inferred relationships, such as a StatefulSet pointing back to Pods it manages, are calculated to facilitate easier graph traversal, even if direct knowledge doesn't exist.

* Each entity has exactly one **owner**, responsible for its creation.  When created, often direct and computed edge relationships are also established.  The owner is responsible to keep entities and edges fresh.

* Each edge has exactly one **owner**, responsible for creating that edge. For instance, an HTTP trace can create proactive edges between Pods it observes without creating the Pods themselves.

* Edges can reference entities that do not yet exist.

* Entities follow a standardized naming convention for their URI, uniquely identifying each entity. The URI can be generated from basic identifying details. For example, a Kubernetes StatefulSet URI format may be `/k8s/clusterName/namespace/k8s.statefulset/name`. A Pod entity created prior to its StatefulSet can still reference it, establishing forward and backward edges. Once the StatefulSet entity is created, it seamlessly integrates into the graph.

* Entities and edges include timestamps: `valid_from`, `valid_to`, and `last_seen`. Attributes have `valid_from` and `valid_to` timestamps.
  * `valid_from` and `valid_to` support historical queries ("time travel"), enabling queries on the graph's state at any previous point in time. Old data can expire to manage storage growth. At any moment, an entity or edge can have only one active validity period (`valid_from` is in the past, `valid_to` is NULL). Multiple historical intervals can exist without overlap.
  * `last_seen` captures the latest timestamp when an owner observed the entity or edge. It allows lazy deletion of stale data without requiring explicit deletion events from the source.  Generally, this field should not be used in queries, but will be used by an expiration detector to update `valid_to` to expire old objects.

### Entities

* Entities can store objects of virtually any type.

* Entities contain attributes, each belonging to a namespace defined by their owner. By convention, only the owning entity updates these attributes.

* Use OpenTelemetry semantic values whenever possible. Non-standard attributes must have unique prefixes to avoid conflicts.

* Entities have defined validity periods (`valid_from`, `valid_to`) to track their lifecycle from creation to deletion. Additionally, a refresh mechanism (`last_seen`) indicates ongoing existence.

* Attributes have defined validity periods for change detection. Due to attribute namespacing, owners easily compare current states against stored attributes, updating attribute sets without overwriting historical data.

### Events

Events are either instanenous or have a time range where the condition was asserted.  For example, a change notification is typically an instanenous event, where at a point in time the change was detected.  An abnormal condition such as an over CPU, over memory, crash-loop-backoff condition, etc. may be an ongoing assertion until that condition clears.

Events are modeled as entities

### Querying

As this graph database is a currated set of object types and their relationships, most recursive queries will need some information on how to get from a node to the data it is wanting, usually a list of a single object like a pod, or a list of characteristics about an object, like what images all containers in all pods in a specific deployment are currently running.

Questions can also be of a temporal nature, such as "yesterday at noon, when an alert was raised about a service, what changed around that time?

In this graph database, these are known as "allowed transitions" that encourage a recursive SQL query to keep moving toward the target data in an efficient manner.

For example, asking the question "what images are pods for a chq.service named 'front-end' currently running?" would have a transition map similar to:

* source chq.service, target type k8s.pod
* chq.service -> any of { k8s.deployment, k8s.statefulset, k8s.daemonset, k8s.cronjob, k8s.job }
* k8s.deployment -> { k8s.replicaset }
* k8s.replicaset -> { k8s.pod }
* k8s.statefulset -> { k8s.pod }
* k8s.daemonset -> { k8s.pod }
* k8s.cronjob -> { k8s.job }
* k8s.job -> { k8s.pod }

Each source and destination may need to have such a map defined.

Alternative approaches are acceptable, such as allowing all transition except certain ones, or requiring some attribute to be present on all discovered objects on the path to the destination.
