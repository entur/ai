# Kafka / Event Streaming

Entur uses **Apache Kafka** on **Aiven** for event streaming. The standard library is **`entur-kafka-spring-starter`** (`org.entur.data:entur-kafka-spring-starter`), providing Spring Boot autoconfiguration with sensible defaults.

For advanced topics, see the [Data Handbook](https://enturas.atlassian.net/wiki/spaces/TD/pages/4962451517/2+-+Data+Streaming+Kafka).

## When to Use Kafka vs REST

| Pattern | Use |
|---------|-----|
| **Kafka** | Event-driven communication, fan-out to multiple consumers, audit trails, high-throughput data pipelines, decoupled producers/consumers |
| **REST** | Synchronous request-response, CRUD operations, client-facing APIs |

## Infrastructure

### Aiven Clusters

Separate clusters per environment, each with internal (VPC-peered) and public endpoints:

| Cluster Enum | Use When | Environment |
|-------------|----------|-------------|
| `AIVEN_TEST_INT` | App runs in `kub-ent-tst` (inside VPC) | Test |
| `AIVEN_PUBLIC_TEST_INT` | App runs locally or outside VPC | Test |
| `AIVEN_TEST_EXT` | External partner access | Test |
| `AIVEN_PROD_INT` | App runs in `kub-ent-prd` (inside VPC) | Production |
| `AIVEN_PUBLIC_PROD_INT` | App runs locally or outside VPC | Production |
| `AIVEN_PROD_EXT` | External partner access | Production |

**Rule**: `*_INT` for apps in the corresponding K8s environment. `*_PUBLIC_*_INT` for local dev or outside VPC. `*_EXT` only for external partners.

### Authentication

All clusters use **SASL/SCRAM-SHA-512** over TLS (`SASL_SSL`). Credentials are provisioned per service and must be stored in **Google Secret Manager** via ExternalSecrets in Helm -- never hardcode them.

### Schema Registry

Each cluster has a **Confluent Schema Registry** for Avro and Protobuf. The URL is auto-resolved from the cluster enum. Auth uses the same SASL credentials as basic auth.

## Dependency Setup

### Gradle

```kotlin
plugins {
    // Required if using Avro schemas (generates Java/Kotlin classes from .avsc files)
    id("com.github.davidmc24.gradle.plugin.avro") version "1.9.1"
}

dependencies {
    implementation("org.springframework.boot:spring-boot-starter")
    implementation("org.entur.data:entur-kafka-spring-starter:<version>")  // check Artifactory for latest
}
```

Published to Entur's [JFrog Artifactory](https://entur2.jfrog.io). See [java.md](java.md#artifactory-jfrog) for repository configuration.

## Configuration

All config is under the `entur.kafka` prefix. The starter autoconfigures both producer and consumer beans.

### Minimal Configuration

```yaml
entur:
  kafka:
    kafkaCluster: "AIVEN_TEST_INT"
    sasl:
      username: ${KAFKA_USERNAME}
      password: ${KAFKA_PASSWORD}
    consumer:
      group: "my-application"
```

### Per-Environment Configuration

Use Spring profiles to select the correct cluster:

```yaml
# application.yml (shared)
entur:
  kafka:
    sasl:
      username: ${KAFKA_USERNAME}
      password: ${KAFKA_PASSWORD}
    consumer:
      group: "my-application"

---
# application-dev.yml / application-tst.yml
entur:
  kafka:
    kafkaCluster: "AIVEN_TEST_INT"

---
# application-prd.yml
entur:
  kafka:
    kafkaCluster: "AIVEN_PROD_INT"

---
# application-local.yml (local development)
entur:
  kafka:
    kafkaCluster: "AIVEN_PUBLIC_TEST_INT"
```

### Serialization Defaults

| Setting | Default | Alternatives |
|---------|---------|-------------|
| Key serializer | `StringSerializer` | `KafkaAvroSerializer` for Avro-typed keys |
| Key deserializer | `StringDeserializer` | `KafkaAvroDeserializer` for Avro-typed keys |
| Value serializer | `KafkaAvroSerializer` | `KafkaProtobufSerializer` for Protobuf |
| Value deserializer | `KafkaAvroDeserializer` | `KafkaProtobufDeserializer` for Protobuf |
| Specific Avro reader | `true` | Set `false` for `GenericRecord` |

### Separate Producer/Consumer Clusters

For producing to a different cluster than consuming from:

```yaml
entur:
  kafka:
    kafkaCluster: "AIVEN_TEST_INT"          # consumer cluster
    producerCluster: "AIVEN_PROD_INT"       # producer cluster (overrides for producers)
    sasl:
      username: ${KAFKA_CONSUMER_USERNAME}
      password: ${KAFKA_CONSUMER_PASSWORD}
      producerUsername: ${KAFKA_PRODUCER_USERNAME}   # separate producer credentials
      producerPassword: ${KAFKA_PRODUCER_PASSWORD}
```

### Toggling Producer/Consumer Beans

```yaml
entur:
  kafka:
    consumer:
      enabled: false   # disable consumer beans (e.g., producer-only service)
    producer:
      enabled: false   # disable producer beans (e.g., consumer-only service)
```

Note: DLT functionality requires the producer to be enabled.

## Producing Messages

### Standard Producer (Avro, String Keys)

Most common pattern -- Avro values with string keys. Inject `EnturKafkaProducer<T>` and call `send()` with topic, key (determines partition), value (Avro `SpecificRecord`), `correlationId()`, success/failure callbacks, and optional custom headers.

The `correlationId` is added as an `X-Correlation-Id` header automatically.

### Avro-Keyed Producer

For Avro-typed keys (both key and value are `SpecificRecordBase`), inject `EnturAvroKeyKafkaProducer<K, V>`.

Requires key serializer config:

```yaml
entur:
  kafka:
    keySerializer: "io.confluent.kafka.serializers.KafkaAvroSerializer"
    keyDeserializer: "io.confluent.kafka.serializers.KafkaAvroDeserializer"
```

### Protobuf Producer

Inject `EnturProtobufKafkaProducer<K, V>` for Protobuf messages.

Requires value serializer config:

```yaml
entur:
  kafka:
    valueSerializer: "io.confluent.kafka.serializers.protobuf.KafkaProtobufSerializer"
```

### Generic Producer

For schemaless topics (legacy use only). Inject `EnturGenericKafkaProducer<K, V>`.

### Producer Tuning Options

```yaml
entur:
  kafka:
    producer:
      acks: "all"                    # all replicas must acknowledge (required for idempotence)
      retries: 2147483647            # max int -- use timeout to control retry duration
      enableIdempotence: true        # exactly-once per partition
      maxInFlightRequests: 5         # max unacknowledged requests (<=5 required for idempotence)
      deliveryTimeoutMs: 120000      # max time for send() to complete
      compressionType: "lz4"         # none, gzip, snappy, lz4, zstd
```

### Transactional Producer

Enable Kafka transactions (exactly-once semantics across partitions):

```yaml
entur:
  kafka:
    producer:
      transactionIdPrefix: "my-app-tx-"   # enables transactions, implicitly sets idempotence
      allowNonTransactional: true          # allow producing outside @Transactional (default true)
```

Use `@Transactional` on methods that produce messages to multiple topics atomically.

## Consuming Messages

### Standard Consumer (Avro)

Use `@KafkaListener` with `containerFactory = "enturListenerFactory"` (the factory with all Entur defaults). Receive key via `@Header(KafkaHeaders.RECEIVED_KEY)` and value via `@Payload`.

**Important**: Always use `containerFactory = "enturListenerFactory"`.

### Avro-Keyed Consumer

Same as standard consumer, but the `@Header(KafkaHeaders.RECEIVED_KEY)` parameter is typed to the Avro key type.

### Protobuf Consumer

Uses a different container factory: `enturSpecificProtobufConsumerFactory`.

```yaml
entur:
  kafka:
    valueDeserializer: "io.confluent.kafka.serializers.protobuf.KafkaProtobufDeserializer"
    consumer:
      specificProtobufMessageValue: "com.example.MyProtoMessage"
```

### GenericRecord Consumer

For schemaless or dynamically typed topics. Receive as `ConsumerRecord<Any, Any>` via `@Payload`. Set `useSpecificAvro: false`:

```yaml
entur:
  kafka:
    consumer:
      useSpecificAvro: false
```

### Consumer Tuning Options

```yaml
entur:
  kafka:
    consumer:
      group: "my-application"         # consumer group ID (required for group coordination)
      offsetReset: "latest"           # "latest" or "earliest" for new consumer groups
      sessionTimeoutMs: 15000         # heartbeat timeout (recommended: 15000-300000)
      maxPollIntervalMs: 300000       # max time between polls before considered failed
      maxPollRecords: 500             # records per poll batch
      enableAutoCommit: true          # automatic offset commits
```

### Standalone Consumer (No Consumer Group)

For consumers that must read all partitions independently. Use `topicPartitions` with `@partitionFinder.allPartitions()` SpEL expression in the `@KafkaListener` annotation instead of `topics`.

```yaml
entur:
  kafka:
    consumer:
      enabled: true
      group:                          # leave null -- no consumer group
      offsetReset: "earliest"         # or "latest"
      enableAutoCommit: false         # no offsets to commit without a group

spring:
  kafka:
    listener:
      ack-mode: "manual"             # disable Spring Kafka's built-in ack handling
```

## Error Handling and Retry

### Non-Blocking Retry with DLT

The starter supports **non-blocking retry** using separate retry topics with exponential backoff. Failed messages move to retry topics, allowing the consumer to continue processing.

```yaml
entur:
  kafka:
    retry:
      enabled: true
      maxAttempts: 3                       # total attempts including original
      initialInterval: 5000                # 5s initial delay
      intervalMultiplier: 5.0              # exponential backoff multiplier
      maxInterval: 125000                  # max 125s between retries
      retryTopics:                         # empty = all topics; or list specific topics
        - "order-events"
      retryTopicsPrefix: ""                # prefix for retry/DLT topic names
      useSamePartition: false              # let Kafka choose partition on retry topics
```

Retry topic naming: `order-events` → `order-events-retry-0`, `order-events-retry-1`, ..., `order-events-dlt`.

### Blocking Retry for Transient Errors

For errors where all messages would fail (e.g., downstream service down), use blocking retries to pause consumption:

```yaml
entur:
  kafka:
    retry:
      enabled: true
      blockingRetryExceptions:
        - "org.springframework.web.client.ResourceAccessException"
      blockingInterval: 1000               # 1s initial blocking delay
      blockingIntervalMultiplier: 5.0
      maxBlockingInterval: 125000
```

### Fatal Exceptions (Skip to DLT)

Exceptions that should never be retried:

```yaml
entur:
  kafka:
    retry:
      fatalExceptions:
        - "com.fasterxml.jackson.core.JsonParseException"
        - "org.entur.myapp.InvalidOrderException"
```

### DLT Handler

Create a `@Component` bean with a handler method that receives the failed message. Log the failure and alert as needed.

```yaml
entur:
  kafka:
    retry:
      dltHandlerBean: "orderDltHandler"
      dltHandlerMethod: "handleDlt"
```

### Manual Retry/DLT with Annotations

If the starter's retry config is insufficient, use Spring Kafka annotations directly (set `entur.kafka.retry.enabled: false`). Use `@RetryableTopic(kafkaTemplate = "enturKafkaTemplate")` on the listener and `@DltHandler` for dead letter handling.

### Custom Retry Exception Logging

Provide a `@Bean` of type `CustomRetryExceptionLogger` to customize logging when messages are retried or sent to DLT. The lambda receives the exception, consumer record, and next destination.

### Handling Deserialization Errors

Messages that fail deserialization never reach the listener. Configure expected Avro classes to route these to the DLT:

```yaml
entur:
  kafka:
    avroSerializableClasses:
      - "org.entur.myapp.OrderEvent"
      - "org.entur.myapp.PaymentEvent"
```

This enables a `DelegatingByTypeSerializer` for DLT publishing. **Keep this list up to date** -- missing types cause the error handler to fail.

### Custom Error Handler

Provide a `@Bean` named `enturCustomErrorHandler` of type `CommonErrorHandler` for fully custom error handling (overrides all defaults including retry topic naming). Use `DefaultErrorHandler` with `DeadLetterPublishingRecoverer` and your preferred backoff strategy.

## Avro Schema Management

Place `.avsc` schema files in `src/main/avro/`. The Gradle Avro plugin generates classes during compilation:

```kotlin
plugins {
    id("com.github.davidmc24.gradle.plugin.avro") version "1.9.1"
}
```

The starter auto-configures the Confluent Schema Registry client based on the selected cluster. Schemas must be backward-compatible by default (enforced at the registry level).

## Testing

### Unit/Integration Test Configuration

```yaml
# application-test.yml
entur:
  kafka:
    bootstrapServer: "localhost:9092"
    schemaRegistryUrl: "mock://testing"
    securityProtocol: "PLAINTEXT"
    sasl:
      mechanism: "PLAIN"
      username: "test"
      password: "test"
    consumer:
      group: "test-group"
```

Override `bootstrapServer` and `schemaRegistryUrl` directly -- takes precedence over `kafkaCluster`.

### Testcontainers with Kafka

Use `KafkaContainer` (from `confluentinc/cp-kafka` image) with `@Testcontainers` and `@DynamicPropertySource`. Override `entur.kafka.bootstrapServer`, `entur.kafka.schemaRegistryUrl` (`"mock://testing"`), and `entur.kafka.securityProtocol` (`"PLAINTEXT"`).

## Redis as Kafka State Store

Redis (Memorystore) is commonly paired with Kafka consumers for deduplication, state caching, and idempotent processing. For Redis infrastructure, see [terraform/modules.md](terraform/modules.md#memorystore-redis). For general Redis patterns, see [java.md](java.md#redis-memorystore) or [go.md](go.md#redis-memorystore).

### Idempotent Consumer (Deduplication)

Kafka provides at-least-once delivery. Use Redis `SET NX EX` to deduplicate. In the `@KafkaListener`, extract the event ID from a header, build a dedup key (`myapp:dedup:{eventId}`), and use `redis.opsForValue().setIfAbsent(key, "1", ttl)`. Skip processing if the key already exists.

**TTL guidance**: Set dedup key TTL to at least the max expected redelivery window. 24h is a safe default; if retry/DLT retries for at most 2h, 4h TTL suffices.

### Consumer State Cache

Cache reference data lookups to avoid repeated DB queries. Use a cache-aside pattern: check Redis first, fall back to repository, then populate cache with TTL. Use `ObjectMapper` for JSON serialization to/from Redis.

### Best Practices for Redis + Kafka

- **Always use TTLs** -- Kafka consumer patterns generate large volumes of keys
- **Handle Redis failures gracefully** -- if Redis is down, either process (risking duplicates) or throw to trigger Kafka retry, based on idempotency requirements
- **Use event ID or Kafka offset as dedup key** -- `{topic}:{partition}:{offset}` is naturally unique
- **ALWAYS use Kafka for event streaming** -- Redis Pub/Sub lacks persistence, consumer groups, and delivery guarantees
- **Namespace keys** with app name to avoid collisions: `myapp:dedup:`, `myapp:cache:`

## Observability

The starter auto-registers **Micrometer/Prometheus** listeners on producer and consumer factories when a `MeterRegistry` bean is present. Standard Kafka client metrics (`kafka_producer_*`, `kafka_consumer_*`) are exposed without additional configuration.

For processing time (`@Timed`) and consumption delay tracking, see [observability.md](observability.md#kafka-consumer-metrics).

## Key Beans Provided

| Bean Name | Type | Condition |
|-----------|------|-----------|
| `enturKafkaTemplate` | `KafkaTemplate<K, V>` | `producer.enabled=true` (default) |
| `kafkaProducer` | `EnturKafkaProducer<T>` | `producer.enabled=true` |
| `kafkaAvroKeyProducer` | `EnturAvroKeyKafkaProducer<K, V>` | `producer.enabled=true` |
| `kafkaGenericProducer` | `EnturGenericKafkaProducer<K, V>` | `producer.enabled=true` |
| `kafkaProtobufProducer` | `EnturProtobufKafkaProducer<K, V>` | `producer.enabled=true` |
| `enturListenerFactory` | `ConcurrentKafkaListenerContainerFactory` | `consumer.enabled=true` (default) |
| `enturSpecificProtobufConsumerFactory` | `ConcurrentKafkaListenerContainerFactory` | `consumer.enabled=true` |
| `enturConsumerFactory` | `ConsumerFactory<K, V>` | `consumer.enabled=true` |
| `partitionFinder` | `PartitionFinder<K, V>` | `consumer.enabled=true` |
| `kafkaTransactionManager` | `KafkaTransactionManager<K, V>` | `transactionIdPrefix` is set |
