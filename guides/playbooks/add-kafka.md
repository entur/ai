# Produce or Consume Kafka Events

Wire a service into Aiven Kafka using the `entur-kafka-spring-starter`.

## Goal

A Spring Boot service that produces and/or consumes Avro (or Protobuf) messages on the correct Aiven cluster for each environment, with retry + DLT configured and observability via Micrometer.

## Prerequisites

- The application is already on the platform (otherwise start with [bootstrap-service.md](bootstrap-service.md)).
- You have decided **Kafka vs REST**: Kafka for event-driven fan-out, audit trails, decoupled producers/consumers; REST for synchronous CRUD. See [entur-kafka-starter.md](../platform/entur-kafka-starter.md#when-to-use-kafka-vs-rest).
- A Kafka topic and SASL credentials have been provisioned (request via `#talk-utviklerplattform` if needed).

## Steps

1. **Add the starter and Avro plugin.** Add `org.entur.data:entur-kafka-spring-starter` and (if using Avro) the `com.github.davidmc24.gradle.plugin.avro` Gradle plugin. See [entur-kafka-starter.md](../platform/entur-kafka-starter.md#dependency-setup).

2. **Select the cluster per environment.** Use Spring profiles: `AIVEN_TEST_INT` in dev/tst pods, `AIVEN_PROD_INT` in prd pods, `AIVEN_PUBLIC_TEST_INT` for local dev. The `*_INT` clusters require the pod to run inside the VPC; `*_PUBLIC_*` is for outside. See [entur-kafka-starter.md](../platform/entur-kafka-starter.md#aiven-clusters).

3. **Wire SASL credentials via Secret Manager + ExternalSecrets.** Add `KAFKA_USERNAME` / `KAFKA_PASSWORD` to `common.secrets` in Helm. Never hardcode SASL credentials. See [common-helm.md](../platform/common-helm.md#secrets-externalsecrets).

4. **Place Avro schemas in `src/main/avro/`.** The Gradle plugin generates classes during compilation. Schemas must be backward-compatible -- the Schema Registry enforces this.

5. **Produce:** inject `EnturKafkaProducer<T>` (for string keys + Avro values) and call `.send()`. Use `correlationId()` -- it becomes the `X-Correlation-Id` header automatically. For exactly-once across partitions, set `transactionIdPrefix` and use `@Transactional`. See [entur-kafka-starter.md](../platform/entur-kafka-starter.md#producing-messages).

6. **Consume:** annotate with `@KafkaListener(containerFactory = "enturListenerFactory", ...)`. Always use the Entur container factory -- it carries retry + DLT defaults. See [entur-kafka-starter.md](../platform/entur-kafka-starter.md#consuming-messages).

7. **Configure retry + DLT.** Enable non-blocking retry (`entur.kafka.retry.enabled: true`) with exponential backoff. Add fatal exceptions (e.g. `JsonParseException`) so poison messages skip retries. Implement a DLT handler bean. See [entur-kafka-starter.md](../platform/entur-kafka-starter.md#error-handling-and-retry).

8. **Add idempotency for at-least-once.** Use Redis `SET NX EX` keyed by event ID to deduplicate redelivered messages. See [add-redis.md](add-redis.md) and [entur-kafka-starter.md](../platform/entur-kafka-starter.md#idempotent-consumer-deduplication).

9. **Wire metrics.** The starter auto-registers Micrometer listeners. Add `@Timed(value = KAFKA_CONSUMER_PROCESS_TIME, ...)` and record consumption delay with `KAFKA_CONSUMER_CONSUME_DELAY`. See [observability.md](../reference/observability.md#kafka-consumer-metrics).

## Verify

- Local: Testcontainers `KafkaContainer` with `schemaRegistryUrl: "mock://testing"` boots and round-trips a message. See [entur-kafka-starter.md](../platform/entur-kafka-starter.md#testing).
- Deployed: Grafana Kafka dashboards show producer/consumer metrics; DLT topic is empty.
- Forced failure (e.g. malformed payload) lands in the DLT topic with retry headers showing the attempt count.

## See also

- Redis for dedup: [add-redis.md](add-redis.md)
- Metrics naming: [observability.md](../reference/observability.md#standard-metric-names)
