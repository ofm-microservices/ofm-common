package resilience

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaRetryQueue publishes failed Kafka records to a service-scoped retry
// topic. The source consumer can commit its original offset after Enqueue
// succeeds, so a failed record cannot block later records in the source topic.
type KafkaRetryQueue struct {
	Brokers     []string
	Group       string
	MaxAttempts int
	Writer      KafkaWriter
}

// KafkaRetryQueueConfig configures the retry-topic worker for one consumer
// group. Retry topics are isolated per group so one service cannot block
// another service's recovery traffic.
type KafkaRetryQueueConfig struct {
	Brokers     []string
	Group       string
	MaxAttempts int
}

// Run consumes retry records, waits until each record is eligible, and
// republishes the original Kafka value to its source topic. Exhausted records
// are written to the group's dead-letter topic and acknowledged on the retry
// topic.
func (q KafkaRetryQueueConfig) Run(ctx context.Context) error {
	if len(q.Brokers) == 0 {
		return fmt.Errorf("retry queue brokers are empty")
	}
	if q.Group == "" {
		return fmt.Errorf("retry queue group is empty")
	}
	if q.MaxAttempts < 1 {
		q.MaxAttempts = DefaultRetryPolicy.MaxAttempts
	}
	r := kafka.NewReader(kafka.ReaderConfig{Brokers: q.Brokers, Topic: RetryTopic(q.Group), GroupID: q.Group + ".retry", MinBytes: 1, MaxBytes: 10e6, MaxWait: 50 * time.Millisecond})
	defer r.Close()
	for {
		message, err := r.FetchMessage(ctx)
		if err != nil {
			return err
		}
		record, err := UnmarshalRetry(message.Value)
		if err != nil {
			if err := q.publishDeadLetter(ctx, message, err); err != nil {
				return err
			}
			if err := r.CommitMessages(ctx, message); err != nil {
				return err
			}
			continue
		}
		if wait := time.Until(record.AvailableAt); wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
		if record.Attempt >= q.MaxAttempts {
			if err := q.publishDeadLetter(ctx, message, fmt.Errorf("retry attempts exhausted: %s", record.LastError)); err != nil {
				return err
			}
		} else {
			writer := &kafka.Writer{Addr: kafka.TCP(q.Brokers...), Topic: record.OriginalTopic, WriteTimeout: 5 * time.Second}
			headers := make([]kafka.Header, 0, len(record.OriginalHeaders)+1)
			for _, header := range record.OriginalHeaders {
				headers = append(headers, kafka.Header{Key: header.Key, Value: header.Value})
			}
			headers = append(headers, kafka.Header{Key: "x-ofm-retry-attempt", Value: []byte(fmt.Sprint(record.Attempt))})
			err := writer.WriteMessages(ctx, kafka.Message{Key: record.OriginalKey, Value: record.OriginalValue, Headers: headers})
			_ = writer.Close()
			if err != nil {
				return err
			}
		}
		if err := r.CommitMessages(ctx, message); err != nil {
			return err
		}
	}
}

func (q KafkaRetryQueueConfig) publishDeadLetter(ctx context.Context, source kafka.Message, cause error) error {
	payload, err := MarshalDLQ(DLQRecord{OriginalKey: source.Key, OriginalValue: source.Value, OriginalTopic: RetryTopic(q.Group), OriginalPartition: source.Partition, OriginalOffset: source.Offset, Attempts: q.MaxAttempts, ErrorClass: fmt.Sprintf("%T", cause), Error: cause.Error(), FailedAt: time.Now().UTC()})
	if err != nil {
		return err
	}
	writer := &kafka.Writer{Addr: kafka.TCP(q.Brokers...), Topic: DeadLetterTopic(q.Group), WriteTimeout: 5 * time.Second}
	err = writer.WriteMessages(ctx, kafka.Message{Key: source.Key, Value: payload})
	_ = writer.Close()
	return err
}

// KafkaWriter is the small Kafka transport contract needed by retry routing.
type KafkaWriter interface {
	WriteMessages(context.Context, ...kafka.Message) error
}

// Enqueue places a failed source record on the retry topic.
func (q KafkaRetryQueue) Enqueue(ctx context.Context, message kafka.Message, sourceTopic string, attempt int, cause error) error {
	if attempt < 1 {
		attempt = 1
	}
	if q.Writer == nil {
		return fmt.Errorf("retry queue writer is nil")
	}
	headers := make([]RetryHeader, 0, len(message.Headers))
	for _, header := range message.Headers {
		headers = append(headers, RetryHeader{Key: header.Key, Value: header.Value})
	}
	payload, err := MarshalRetry(RetryEnvelope{
		OriginalKey: message.Key, OriginalValue: message.Value,
		OriginalTopic: sourceTopic, OriginalPartition: message.Partition,
		OriginalOffset: message.Offset, OriginalHeaders: headers, Attempt: attempt,
		AvailableAt: time.Now().UTC().Add(BackoffDelay(DefaultRetryPolicy, attempt)),
		LastError:   errorString(cause),
	})
	if err != nil {
		return err
	}
	return q.Writer.WriteMessages(ctx, kafka.Message{Key: message.Key, Value: payload})
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
