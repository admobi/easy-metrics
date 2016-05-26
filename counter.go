package easy-metrics

import (
	"strconv"
	"sync/atomic"
)

// Counter is a cumulative metric that represents a single numerical value that only ever goes up.
// Satsfies Metric interface
type Counter struct {
	name  string
	value uint64
}

// NewCounter returns new counter that satsfies Metric interface
func NewCounter(name string) *Counter {
	return &Counter{name: name}
}

// Get returns counter value
func (c *Counter) Get() interface{} {
	return atomic.LoadUint64(&c.value)
}

// Add adds delta to counter value.
func (c *Counter) Add(delta uint64) {
	atomic.AddUint64(&c.value, delta)
}

// Inc increases counter value by 1
func (c *Counter) Inc() {
	c.Add(1)
}

// String returns formated representation of counter value.
func (c *Counter) String() string {
	return strconv.FormatUint(c.Get().(uint64), 10)
}

// Name returns metric name.
func (c *Counter) Name() string {
	return c.name
}

func (c *Counter) copy() Metric {
	return &Counter{value: atomic.LoadUint64(&c.value), name: c.name}
}

func (c *Counter) flush() {
	atomic.StoreUint64(&c.value, 0)
}
