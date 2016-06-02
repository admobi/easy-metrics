package metrics

import (
	"sync"
	"time"
)

// Registry is an abstract type for metric countainer
type Registry interface {
	AddMetrics(metrics ...Metric) error
	GetMetricByName(name string) (Metric, error)
	GetMetrics() map[string]Metric
}

// registryMap is a global container that holds all created registries
var registryMap struct {
	sync.Mutex
	r map[string]Registry
}

func init() {
	registryMap.r = make(map[string]Registry)
}

// GetRegistries returns all registries
func GetRegistries() map[string]Registry {
	registryMap.Lock()
	defer registryMap.Unlock()
	return registryMap.r
}

// GetRegistryByName returns Registry by given name
func GetRegistryByName(name string) (Registry, error) {
	if len(name) == 0 {
		return nil, ErrEmptyRegistryName{}
	}

	registryMap.Lock()
	defer registryMap.Unlock()
	if _, ok := registryMap.r[name]; !ok {
		return nil, ErrRegistryUnknown(name)
	}
	return registryMap.r[name], nil
}

// DefaultRegistry its a plain container for metrics.
// It just keeps metric
type DefaultRegistry struct {
	sync.Mutex
	metrics map[string]Metric
	// Ordered map for metric's keys
	orderedKeys []string
}

// NewRegistry creates a new registry and adds it into the registry map
func NewRegistry(name string) (Registry, error) {
	if len(name) == 0 {
		return nil, ErrEmptyRegistryName{}
	}

	registryMap.Lock()
	defer registryMap.Unlock()

	if _, ok := registryMap.r[name]; ok {
		return nil, ErrRegistryExists(name)
	}

	registryMap.r[name] = &DefaultRegistry{
		metrics: make(map[string]Metric),
	}

	return registryMap.r[name], nil
}

// AddMetrics adds one or more metrics into registry
func (r *DefaultRegistry) AddMetrics(metrics ...Metric) error {
	r.Lock()
	defer r.Unlock()
	for _, m := range metrics {
		name := m.Name()

		if len(name) == 0 {
			return ErrEmptyMetricName{}
		}

		if _, ok := r.metrics[name]; ok {
			return ErrMetricExists(name)
		}

		r.metrics[name] = m
		r.orderedKeys = append(r.orderedKeys, name)
		// sort.Strings(r.sortedMap)
	}
	return nil
}

// GetMetricByName returns metric by given name
func (r *DefaultRegistry) GetMetricByName(name string) (Metric, error) {
	if len(name) == 0 {
		return nil, ErrEmptyMetricName{}
	}
	r.Lock()
	defer r.Unlock()

	if _, ok := r.metrics[name]; !ok {
		return nil, ErrMetricUnknown(name)
	}

	return r.metrics[name], nil
}

// GetMetrics returns map of registred metrics
func (r *DefaultRegistry) GetMetrics() map[string]Metric {
	r.Lock()
	defer r.Unlock()
	return r.metrics
}

// Tracker is an abstract type for countainer with metrics snaphshot
// Implements Registry interface
type Tracker interface {
	Registry
	GetSnapshots() []Snapshot
}

// TrackRegistry is a registry that can stores the pool of snapshoted metrics.
type TrackRegistry struct {
	timer    *time.Ticker
	duration time.Duration
	// Metric snapshots container
	buf []Snapshot
	DefaultRegistry
}

// Snapshot is a struct for snapshoted metrics
// It stores snapshot timestamp and map of metrics
type Snapshot struct {
	// Timestamp of metrics archivation
	t time.Time
	// Archived metrics
	data map[string]Metric
}

// GetTimestamp returns timestamp of metrics snapshot
func (am *Snapshot) GetTimestamp() time.Time {
	return am.t
}

// GetMetricByName returns Metric by given name
func (am *Snapshot) GetMetricByName(name string) (Metric, error) {
	if _, ok := am.data[name]; !ok {
		return nil, ErrMetricUnknown(name)
	}
	return am.data[name], nil
}

// GetMetrics returns all metrics from snapshot
func (am *Snapshot) GetMetrics() map[string]Metric {
	return am.data
}

// NewTrackRegistry creates a new TrackRegistry and adds it into the registry map.
// It makes the snapshots of metric on each interval and keeps it in pool with specified capacity.
// If align is set to true, metric's archiving will be align by interval duration.
// For examaple:
//
//
//      // If application will be started at 12:10:13
//
//      // this registry will create snaphosts at 13:10:13, 14:10:13, 15:10:13 etc.
//      NewTrackRegistry("stat per minute", 10, time.Hour, false)
//
//      // but this will create snaphosts at 13:00:00, 14:00:00, 15:00:00 etc.
//      NewTrackRegistry("stat per minute", 10, time.Hour, true)
//
//
func NewTrackRegistry(name string, capacity int, interval time.Duration, align bool) (Tracker, error) {
	if len(name) == 0 {
		return nil, ErrEmptyRegistryName{}
	}

	registryMap.Lock()
	defer registryMap.Unlock()

	if _, ok := registryMap.r[name]; ok {
		return nil, ErrRegistryExists(name)
	}

	trackReg := &TrackRegistry{
		buf:      make([]Snapshot, 0, capacity),
		duration: interval,
	}

	trackReg.metrics = make(map[string]Metric)
	registryMap.r[name] = trackReg

	// align snaphshots creation by interval
	if align {
		ds := time.Second * time.Duration((time.Now().Truncate(interval).Add(interval).Unix() - time.Now().Unix()))
		go time.AfterFunc(
			ds,
			func() {
				trackReg.makeSnapshot()
				trackReg.timer = time.NewTicker(interval)
				go trackReg.startTimer()
			},
		)
	} else {
		trackReg.timer = time.NewTicker(interval)
		go trackReg.startTimer()
	}

	return registryMap.r[name].(Tracker), nil
}

// GetSnapshots returns slice of swaped metrics
func (r *TrackRegistry) GetSnapshots() []Snapshot {
	r.Lock()
	sn := append([]Snapshot{}, r.buf...)
	r.Unlock()
	return sn
}

// Timer for metrics swapping
func (r *TrackRegistry) startTimer() {
	for {
		select {
		case <-r.timer.C:
			r.makeSnapshot()
		}
	}
}

// Stores the snapshot of current metrics into the swap buffer
// and starts new metrics
func (r *TrackRegistry) makeSnapshot() {
	r.Lock()
	defer r.Unlock()

	r.shiftSlice()

	if len(r.buf) == 0 || len(r.buf) < cap(r.buf) {
		swmetric := Snapshot{}
		swmetric.data = make(map[string]Metric)
		r.buf = append(r.buf, swmetric)
	}
	r.buf[0].data = copyMetrics(r.metrics)
	r.buf[0].t = time.Now().UTC()

	for n := range r.metrics {
		r.metrics[n].flush()
	}
}

// Shift snapshots slice by 1 and pops oldest snapshot
func (r *TrackRegistry) shiftSlice() {
	for i := len(r.buf) - 1; i > 0; i-- {
		r.buf[i] = r.buf[i-1]
		r.buf[i-1].t = r.buf[i-1].t.Add(-r.duration)
	}
}

func copyMetrics(src map[string]Metric) map[string]Metric {
	ret := make(map[string]Metric)
	for name, m := range src {
		ret[name] = m.copy()
	}
	return ret
}
