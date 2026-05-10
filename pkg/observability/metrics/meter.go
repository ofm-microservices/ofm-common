package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type prometheusMeter struct {
	registry *prometheus.Registry

	httpRequests      *prometheus.CounterVec
	httpDuration      *prometheus.HistogramVec
	httpRequestBytes  *prometheus.HistogramVec
	httpResponseBytes *prometheus.HistogramVec
	httpInFlight      prometheus.Gauge

	grpcServerRequests *prometheus.CounterVec
	grpcServerDuration *prometheus.HistogramVec
	grpcClientRequests *prometheus.CounterVec
	grpcClientDuration *prometheus.HistogramVec

	natsReceived          *prometheus.CounterVec
	natsProcessed         *prometheus.CounterVec
	natsProcessingTime    *prometheus.HistogramVec
	natsAck               *prometheus.CounterVec
	natsNak               *prometheus.CounterVec
	natsFetchErrors       *prometheus.CounterVec
	natsConsumerPending   *prometheus.GaugeVec
	natsConsumerBatchSize *prometheus.GaugeVec

	dbOperations    *prometheus.CounterVec
	dbDuration      *prometheus.HistogramVec
	dbTransactions  *prometheus.CounterVec
	redisOperations *prometheus.CounterVec
	redisDuration   *prometheus.HistogramVec
	redisHits       *prometheus.CounterVec
	redisMisses     *prometheus.CounterVec
	objectOps       *prometheus.CounterVec
	objectDuration  *prometheus.HistogramVec

	sagaStarted               *prometheus.CounterVec
	sagaCompleted             *prometheus.CounterVec
	sagaFailed                *prometheus.CounterVec
	sagaCompensationStarted   *prometheus.CounterVec
	sagaCompensationCompleted *prometheus.CounterVec
	sagaCompensationFailed    *prometheus.CounterVec
	sagaStepDuration          *prometheus.HistogramVec
	sagaActiveSessions        *prometheus.GaugeVec
}

// New constructs a Prometheus-backed metrics recorder for one service process.
func New(service, env string) Meter {
	labels := prometheus.Labels{"service": service, "env": env}
	reg := prometheus.NewRegistry()
	m := &prometheusMeter{registry: reg}

	m.httpRequests = counter(reg, "ofm_http_requests_total", "HTTP requests by route and status class.", labels, []string{"method", "route", "status_class"})
	m.httpDuration = histogram(reg, "ofm_http_request_duration_seconds", "HTTP request duration.", labels, []string{"method", "route", "status_class"}, transportBuckets)
	m.httpRequestBytes = histogram(reg, "ofm_http_request_size_bytes", "HTTP request size.", labels, []string{"method", "route"}, sizeBuckets)
	m.httpResponseBytes = histogram(reg, "ofm_http_response_size_bytes", "HTTP response size.", labels, []string{"method", "route"}, sizeBuckets)
	m.httpInFlight = gauge(reg, "ofm_http_in_flight_requests", "HTTP requests currently in flight.", labels)

	m.grpcServerRequests = counter(reg, "ofm_grpc_server_requests_total", "gRPC server requests.", labels, []string{"grpc_service", "grpc_method", "grpc_code"})
	m.grpcServerDuration = histogram(reg, "ofm_grpc_server_duration_seconds", "gRPC server request duration.", labels, []string{"grpc_service", "grpc_method", "grpc_code"}, transportBuckets)
	m.grpcClientRequests = counter(reg, "ofm_grpc_client_requests_total", "gRPC client requests.", labels, []string{"grpc_service", "grpc_method", "grpc_code"})
	m.grpcClientDuration = histogram(reg, "ofm_grpc_client_duration_seconds", "gRPC client request duration.", labels, []string{"grpc_service", "grpc_method", "grpc_code"}, transportBuckets)

	m.natsReceived = counter(reg, "ofm_nats_messages_received_total", "NATS messages received.", labels, []string{"stream", "subject", "durable"})
	m.natsProcessed = counter(reg, "ofm_nats_messages_processed_total", "NATS messages processed.", labels, []string{"stream", "subject", "durable", "status"})
	m.natsProcessingTime = histogram(reg, "ofm_nats_message_processing_duration_seconds", "NATS message processing duration.", labels, []string{"stream", "subject", "durable", "status"}, transportBuckets)
	m.natsAck = counter(reg, "ofm_nats_message_ack_total", "NATS message acknowledgements.", labels, []string{"stream", "subject", "durable"})
	m.natsNak = counter(reg, "ofm_nats_message_nak_total", "NATS negative acknowledgements.", labels, []string{"stream", "subject", "durable"})
	m.natsFetchErrors = counter(reg, "ofm_nats_fetch_errors_total", "NATS pull fetch errors.", labels, []string{"stream", "subject", "durable"})
	m.natsConsumerPending = gaugeVec(reg, "ofm_nats_consumer_pending_messages", "NATS consumer pending messages.", labels, []string{"stream", "subject", "durable"})
	m.natsConsumerBatchSize = gaugeVec(reg, "ofm_nats_consumer_batch_size", "NATS consumer batch size.", labels, []string{"stream", "subject", "durable"})

	m.dbOperations = counter(reg, "ofm_db_operations_total", "Database operations.", labels, []string{"store", "operation", "table", "status"})
	m.dbDuration = histogram(reg, "ofm_db_operation_duration_seconds", "Database operation duration.", labels, []string{"store", "operation", "table", "status"}, storageBuckets)
	m.dbTransactions = counter(reg, "ofm_db_transactions_total", "Database transactions.", labels, []string{"store", "status"})
	m.redisOperations = counter(reg, "ofm_redis_operations_total", "Redis operations.", labels, []string{"operation", "keyspace", "status"})
	m.redisDuration = histogram(reg, "ofm_redis_operation_duration_seconds", "Redis operation duration.", labels, []string{"operation", "keyspace", "status"}, storageBuckets)
	m.redisHits = counter(reg, "ofm_redis_cache_hits_total", "Redis cache hits.", labels, []string{"keyspace"})
	m.redisMisses = counter(reg, "ofm_redis_cache_misses_total", "Redis cache misses.", labels, []string{"keyspace"})
	m.objectOps = counter(reg, "ofm_object_storage_operations_total", "Object storage operations.", labels, []string{"operation", "bucket", "status"})
	m.objectDuration = histogram(reg, "ofm_object_storage_operation_duration_seconds", "Object storage operation duration.", labels, []string{"operation", "bucket", "status"}, storageBuckets)

	m.sagaStarted = counter(reg, "ofm_saga_started_total", "Saga sessions started.", labels, []string{"saga"})
	m.sagaCompleted = counter(reg, "ofm_saga_completed_total", "Saga sessions completed.", labels, []string{"saga"})
	m.sagaFailed = counter(reg, "ofm_saga_failed_total", "Saga sessions failed.", labels, []string{"saga"})
	m.sagaCompensationStarted = counter(reg, "ofm_saga_compensation_started_total", "Saga compensation attempts started.", labels, []string{"saga"})
	m.sagaCompensationCompleted = counter(reg, "ofm_saga_compensation_completed_total", "Saga compensation attempts completed.", labels, []string{"saga"})
	m.sagaCompensationFailed = counter(reg, "ofm_saga_compensation_failed_total", "Saga compensation attempts failed.", labels, []string{"saga"})
	m.sagaStepDuration = histogram(reg, "ofm_saga_step_duration_seconds", "Saga step duration.", labels, []string{"saga", "step", "status"}, sagaBuckets)
	m.sagaActiveSessions = gaugeVec(reg, "ofm_saga_active_sessions", "Active saga sessions.", labels, []string{"saga"})

	return m
}

func (m *prometheusMeter) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
	})
}

func (m *prometheusMeter) Enabled() bool    { return true }
func (m *prometheusMeter) IncHTTPInFlight() { m.httpInFlight.Inc() }
func (m *prometheusMeter) DecHTTPInFlight() { m.httpInFlight.Dec() }

func (m *prometheusMeter) ObserveHTTP(method, route, statusClass string, duration time.Duration, requestBytes, responseBytes int) {
	m.httpRequests.WithLabelValues(method, route, statusClass).Inc()
	m.httpDuration.WithLabelValues(method, route, statusClass).Observe(duration.Seconds())
	m.httpRequestBytes.WithLabelValues(method, route).Observe(float64(requestBytes))
	m.httpResponseBytes.WithLabelValues(method, route).Observe(float64(responseBytes))
}

func (m *prometheusMeter) ObserveGRPCServer(service, method, code string, duration time.Duration) {
	m.grpcServerRequests.WithLabelValues(service, method, code).Inc()
	m.grpcServerDuration.WithLabelValues(service, method, code).Observe(duration.Seconds())
}

func (m *prometheusMeter) ObserveGRPCClient(service, method, code string, duration time.Duration) {
	m.grpcClientRequests.WithLabelValues(service, method, code).Inc()
	m.grpcClientDuration.WithLabelValues(service, method, code).Observe(duration.Seconds())
}

func (m *prometheusMeter) IncNATSReceived(stream, subject, durable string) {
	m.natsReceived.WithLabelValues(stream, subject, durable).Inc()
}

func (m *prometheusMeter) ObserveNATSProcessed(stream, subject, durable, status string, duration time.Duration) {
	m.natsProcessed.WithLabelValues(stream, subject, durable, status).Inc()
	m.natsProcessingTime.WithLabelValues(stream, subject, durable, status).Observe(duration.Seconds())
}

func (m *prometheusMeter) IncNATSAck(stream, subject, durable string) {
	m.natsAck.WithLabelValues(stream, subject, durable).Inc()
}

func (m *prometheusMeter) IncNATSNak(stream, subject, durable string) {
	m.natsNak.WithLabelValues(stream, subject, durable).Inc()
}

func (m *prometheusMeter) IncNATSFetchError(stream, subject, durable string) {
	m.natsFetchErrors.WithLabelValues(stream, subject, durable).Inc()
}

func (m *prometheusMeter) SetNATSPending(stream, subject, durable string, pending int) {
	m.natsConsumerPending.WithLabelValues(stream, subject, durable).Set(float64(pending))
}

func (m *prometheusMeter) SetNATSBatchSize(stream, subject, durable string, size int) {
	m.natsConsumerBatchSize.WithLabelValues(stream, subject, durable).Set(float64(size))
}

func (m *prometheusMeter) ObserveDB(store, operation, table, status string, duration time.Duration) {
	m.dbOperations.WithLabelValues(store, operation, table, status).Inc()
	m.dbDuration.WithLabelValues(store, operation, table, status).Observe(duration.Seconds())
}

func (m *prometheusMeter) IncDBTransaction(store, status string) {
	m.dbTransactions.WithLabelValues(store, status).Inc()
}

func (m *prometheusMeter) ObserveRedis(operation, keyspace, status string, duration time.Duration) {
	m.redisOperations.WithLabelValues(operation, keyspace, status).Inc()
	m.redisDuration.WithLabelValues(operation, keyspace, status).Observe(duration.Seconds())
}

func (m *prometheusMeter) IncRedisHit(keyspace string) { m.redisHits.WithLabelValues(keyspace).Inc() }
func (m *prometheusMeter) IncRedisMiss(keyspace string) {
	m.redisMisses.WithLabelValues(keyspace).Inc()
}

func (m *prometheusMeter) ObserveObjectStorage(operation, bucket, status string, duration time.Duration) {
	m.objectOps.WithLabelValues(operation, bucket, status).Inc()
	m.objectDuration.WithLabelValues(operation, bucket, status).Observe(duration.Seconds())
}

func (m *prometheusMeter) IncSagaStarted(saga string)   { m.sagaStarted.WithLabelValues(saga).Inc() }
func (m *prometheusMeter) IncSagaCompleted(saga string) { m.sagaCompleted.WithLabelValues(saga).Inc() }
func (m *prometheusMeter) IncSagaFailed(saga string)    { m.sagaFailed.WithLabelValues(saga).Inc() }
func (m *prometheusMeter) IncSagaCompensationStarted(saga string) {
	m.sagaCompensationStarted.WithLabelValues(saga).Inc()
}
func (m *prometheusMeter) IncSagaCompensationCompleted(saga string) {
	m.sagaCompensationCompleted.WithLabelValues(saga).Inc()
}
func (m *prometheusMeter) IncSagaCompensationFailed(saga string) {
	m.sagaCompensationFailed.WithLabelValues(saga).Inc()
}
func (m *prometheusMeter) ObserveSagaStep(saga, step, status string, duration time.Duration) {
	m.sagaStepDuration.WithLabelValues(saga, step, status).Observe(duration.Seconds())
}
func (m *prometheusMeter) SetSagaActiveSessions(saga string, active int) {
	m.sagaActiveSessions.WithLabelValues(saga).Set(float64(active))
}
