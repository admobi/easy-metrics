package metrics

import (
	"math"
	"strconv"
	"sync/atomic"
)

// Gauge is a metric that represents a single float64 value that can arbitrarily go up and down.
// Satsfies Metric interface.
type Gauge struct {
	name  string
	value uint64
}

// NewGauge returns new gauge metric that satsfies Metric interface.
func NewGauge(name string) *Gauge {
	return &Gauge{name: name}
}

// Get returns gauge value.
func (g *Gauge) Get() interface{} {
	return math.Float64frombits(atomic.LoadUint64(&g.value))
}

// Add adds delta to gauge value.
func (g *Gauge) Add(delta float64) {
	for {
		cur := atomic.LoadUint64(&g.value)
		curVal := math.Float64frombits(cur)
		nxtVal := curVal + delta
		nxt := math.Float64bits(nxtVal)
		if atomic.CompareAndSwapUint64(&g.value, cur, nxt) {
			return
		}
	}
}

// Set sets gauge value to value.
func (g *Gauge) Set(value float64) {
	atomic.StoreUint64(&g.value, math.Float64bits(value))
}

// Sub substarcts delta from gauge value.
func (g *Gauge) Sub(delta float64) {
	g.Add(-delta)
}

// String returns formated representation of gauge value.
func (g *Gauge) String() string {
	return strconv.FormatFloat(g.Get().(float64), 'g', -1, 64)
}

// Name returns metric name.
func (g *Gauge) Name() string {
	return g.name
}

// Returns copy of gauge. It needs for snapshots.
func (g *Gauge) copy() Metric {
	return &Gauge{value: atomic.LoadUint64(&g.value), name: g.name}
}

// Flush gauge value. It needs for snapshots.
func (g *Gauge) flush() {
	atomic.StoreUint64(&g.value, 0)
}
