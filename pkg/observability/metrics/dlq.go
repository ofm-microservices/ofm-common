package metrics

// DLQRecorder is implemented by meters that expose dead-letter counters.
type DLQRecorder interface {
	IncKafkaDLQ(topic string)
}

// IncKafkaDLQ records a successful write to a Kafka dead-letter topic when the
// configured process meter supports the DLQ metric.
func IncKafkaDLQ(topic string) {
	if recorder, ok := Global().(DLQRecorder); ok {
		recorder.IncKafkaDLQ(topic)
	}
}
