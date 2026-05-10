package metrics

import (
	"net/http"
	"time"
)

type noopMeter struct{}

// NewNoop returns a disabled meter that drops all observations.
func NewNoop() Meter { return noopMeter{} }

func (noopMeter) Handler() http.Handler                                              { return http.NotFoundHandler() }
func (noopMeter) Enabled() bool                                                      { return false }
func (noopMeter) IncHTTPInFlight()                                                   {}
func (noopMeter) DecHTTPInFlight()                                                   {}
func (noopMeter) ObserveHTTP(string, string, string, time.Duration, int, int)        {}
func (noopMeter) ObserveGRPCServer(string, string, string, time.Duration)            {}
func (noopMeter) ObserveGRPCClient(string, string, string, time.Duration)            {}
func (noopMeter) IncNATSReceived(string, string, string)                             {}
func (noopMeter) ObserveNATSProcessed(string, string, string, string, time.Duration) {}
func (noopMeter) IncNATSAck(string, string, string)                                  {}
func (noopMeter) IncNATSNak(string, string, string)                                  {}
func (noopMeter) IncNATSFetchError(string, string, string)                           {}
func (noopMeter) SetNATSPending(string, string, string, int)                         {}
func (noopMeter) SetNATSBatchSize(string, string, string, int)                       {}
func (noopMeter) ObserveDB(string, string, string, string, time.Duration)            {}
func (noopMeter) IncDBTransaction(string, string)                                    {}
func (noopMeter) ObserveRedis(string, string, string, time.Duration)                 {}
func (noopMeter) IncRedisHit(string)                                                 {}
func (noopMeter) IncRedisMiss(string)                                                {}
func (noopMeter) ObserveObjectStorage(string, string, string, time.Duration)         {}
func (noopMeter) IncSagaStarted(string)                                              {}
func (noopMeter) IncSagaCompleted(string)                                            {}
func (noopMeter) IncSagaFailed(string)                                               {}
func (noopMeter) IncSagaCompensationStarted(string)                                  {}
func (noopMeter) IncSagaCompensationCompleted(string)                                {}
func (noopMeter) IncSagaCompensationFailed(string)                                   {}
func (noopMeter) ObserveSagaStep(string, string, string, time.Duration)              {}
func (noopMeter) SetSagaActiveSessions(string, int)                                  {}
