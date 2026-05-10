package metrics

import "sync"

var (
	globalMeterMu sync.RWMutex
	globalMeter   Meter = NewNoop()
)

// SetGlobal installs the process-wide metrics recorder used by lightweight
// instrumentation in services.
func SetGlobal(meter Meter) {
	globalMeterMu.Lock()
	defer globalMeterMu.Unlock()

	if meter == nil {
		globalMeter = NewNoop()
		return
	}
	globalMeter = meter
}

// Global returns the process-wide metrics recorder.
func Global() Meter {
	globalMeterMu.RLock()
	defer globalMeterMu.RUnlock()

	if globalMeter == nil {
		return NewNoop()
	}
	return globalMeter
}
