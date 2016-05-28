// Package metrics provide easy to use, stand by metrics and exposes it via HTTP.
// It can create metric's snapshots each defined duration and store in a pool.
// That useful for fast monitoring of current performance without necessity of export it to external applications.
package metrics

// Metric is an abstract type of metric
type Metric interface {
	// Get returns a metric value
	Get() interface{}
	// String returns formatted value of metric
	String() string
	// Name returns metric name
	Name() string
	// fluhes the metric values
	flush()
	// copy returns a new metric object with current values
	copy() Metric
}
