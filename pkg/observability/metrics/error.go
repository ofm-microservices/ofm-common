package metrics

import "errors"

// ErrNilMeter is returned when metrics HTTP serving is requested without a
// metrics registry.
var ErrNilMeter = errors.New("metrics meter is nil")
