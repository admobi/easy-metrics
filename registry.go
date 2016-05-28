package metrics

import (
	"fmt"
	"sync"
	"time"
)

// Registry is an abstract type for metric countainer
type Registry interface {
	AddMetrics(metrics ...Metric) error
	dump() string
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

func DumpRegistry(name string) (string, error) {
	if len(name) == 0 {
		return "", ErrEmptyRegistryName
	}

	registryMap.Lock()
	defer registryMap.Unlock()

	if _, ok := registryMap.r[name]; !ok {
		return "", ErrRegistryExists(name)
	}
	ret := fmt.Sprintf("%s\n==============\n\n", name)
	ret += registryMap.r[name].dump()
	return ret, nil
}

// GetRegistries returns all registries
func GetRegistries() map[string]Registry {
	registryMap.Lock()
	defer registryMap.Unlock()
	return registryMap.r
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
		return nil, ErrEmptyRegistryName
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
			return ErrEmptyMetricName
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
		return nil, ErrEmptyMetricName
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

func (r *DefaultRegistry) dump() string {
	return dumpMetrics(r.metrics, r.orderedKeys)
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

// GetMetric returns Metric by name
func (am *Snapshot) GetMetric(name string) Metric {
	return am.data[name]
}

// NewTrackRegistry creates a new TrackRegistry and adds it into the registry map.
// It makes the snapshots of  metric on each interval and keeps it in pool with specified capacity.
// If align is set to true, metric's archiving will be align by interval duration.
// For examaple:
//
//
//      // If application will be started at 12:10:13
//
//      // this registry will swap metrics at 12:11:13, 12:12:13, 12:12:13 etc.
//      r, _ := NewTrackRegistry("stat per minute", 10, time.Minute, false)
//
//      // but this will swap metrics at 12:11:00, 12:12:00, 12:12:00 etc.
//      r, _ := NewTrackRegistry("stat per minute", 10, time.Minute, true)
//
//
func NewTrackRegistry(name string, capacity int, interval time.Duration, align bool) (Tracker, error) {
	if len(name) == 0 {
		return nil, ErrEmptyRegistryName
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
	defer r.Unlock()
	return r.buf
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
	if len(r.buf) == 0 || len(r.buf) < cap(r.buf) {
		swmetric := Snapshot{}
		swmetric.data = make(map[string]Metric)
		r.buf = append(r.buf, swmetric)
	}
	shiftSlice(r.buf, r.duration)
	r.buf[0].data = copyMetrics(r.metrics)
	r.buf[0].t = time.Now().UTC()

	for n := range r.metrics {
		r.metrics[n].flush()
	}
}

func (r *TrackRegistry) dump() string {
	r.Lock()
	defer r.Unlock()
	ret := "Current:\n----------\n"
	ret += dumpMetrics(r.metrics, r.orderedKeys)
	ret += "\n"
	ret += "Last:\n----------"
	for _, v := range r.buf {
		ret += fmt.Sprintf("\n[%s]\n", v.t.Format("2006-01-02 15:04:05"))
		ret += dumpMetrics(v.data, r.orderedKeys)
	}
	return ret
}

func shiftSlice(buf []Snapshot, step time.Duration) {
	for i := len(buf) - 1; i > 0; i-- {
		buf[i] = buf[i-1]
		buf[i-1].t = buf[i-1].t.Add(-step)
	}
}

func copyMetrics(src map[string]Metric) map[string]Metric {
	ret := make(map[string]Metric)
	for name, m := range src {
		ret[name] = m.copy()
	}
	return ret
}

func dumpMetrics(m map[string]Metric, sorted []string) (ret string) {
	for _, k := range sorted {
		ret += fmt.Sprintf("%s: %v\n", k, m[k])
	}
	return ret
}
