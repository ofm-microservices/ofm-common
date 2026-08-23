package metrics

import (
	"net/http"
	"time"
)

// Meter records service metrics using stable OFM metric names and labels.
type Meter interface {
	Handler() http.Handler
	Enabled() bool
	IncHTTPInFlight()
	DecHTTPInFlight()
	ObserveHTTP(method, route, statusClass string, duration time.Duration, requestBytes, responseBytes int)
	ObserveGRPCServer(service, method, code string, duration time.Duration)
	ObserveGRPCClient(service, method, code string, duration time.Duration)
	ObserveCircuitBreaker(dependency, state string)
	IncNATSReceived(stream, subject, durable string)
	ObserveNATSProcessed(stream, subject, durable, status string, duration time.Duration)
	IncNATSAck(stream, subject, durable string)
	IncNATSNak(stream, subject, durable string)
	IncNATSFetchError(stream, subject, durable string)
	SetNATSPending(stream, subject, durable string, pending int)
	SetNATSBatchSize(stream, subject, durable string, size int)
	IncKafkaReceived(topic string)
	ObserveKafkaProcessed(topic, status string, duration time.Duration)
	IncKafkaRetry(topic string)
	IncKafkaError(topic string)
	ObserveDB(store, operation, table, status string, duration time.Duration)
	IncDBTransaction(store, status string)
	ObserveRedis(operation, keyspace, status string, duration time.Duration)
	IncRedisHit(keyspace string)
	IncRedisMiss(keyspace string)
	ObserveObjectStorage(operation, bucket, status string, duration time.Duration)
	IncSagaStarted(saga string)
	IncSagaCompleted(saga string)
	IncSagaFailed(saga string)
	IncSagaCompensationStarted(saga string)
	IncSagaCompensationCompleted(saga string)
	IncSagaCompensationFailed(saga string)
	ObserveSagaStep(saga, step, status string, duration time.Duration)
	SetSagaActiveSessions(saga string, active int)
}
