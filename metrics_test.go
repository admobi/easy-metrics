package metrics_test

import (
	"testing"
	"time"

	"github.com/admobi/easy-metrics"
)

func BenchmarkCounter(b *testing.B) {
	m1 := metrics.NewCounter("Total_connections")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m1.Inc()
	}
}

func BenchmarkGauge(b *testing.B) {
	m1 := metrics.NewGauge("Total connections")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m1.Add(1e-13)
	}
}

func TestNewTrackRegistry(t *testing.T) {
	_, err := metrics.NewTrackRegistry("newswap", 10, time.Second*100, true)
	if err != nil {
		t.Errorf("unable to create swap registry, %v", err)
	}

	_, err = metrics.NewTrackRegistry("", 10, time.Second*100, false)

	if err == nil {
		t.Error("empty registry name, should be error but got nil")
	}

	_, err = metrics.NewTrackRegistry("newswap", 10, time.Second*200, false)

	if err == nil {
		t.Error("existing registry name, should be error but got nil")
	}

	switch err.(type) {
	case metrics.ErrRegistryExists:
	default:
		t.Errorf("undefined error %v", err)
	}
}

func TestRegisterMetric(t *testing.T) {
	rg, _ := metrics.NewTrackRegistry("newcounter", 10, time.Second*100, false)
	m1 := metrics.NewCounter("name1")
	err := rg.AddMetrics(m1)
	if err != nil {
		t.Errorf("unable to register metric, %v", err)
	}
	m2 := metrics.NewCounter("name1")
	err = rg.AddMetrics(m2)
	if err == nil {
		t.Error("existing metric name, should be error on register but got nil")
	}

	switch err.(type) {
	case metrics.ErrMetricExists:
	default:
		t.Errorf("undefined error %v", err)
	}
}

func TestDefaultCounter(t *testing.T) {
	rg, _ := metrics.NewTrackRegistry("testcounters", 10, time.Second*100, false)
	m1 := metrics.NewCounter("Total connections")
	rg.AddMetrics(m1)
	m2 := metrics.NewCounter("Total connections 2")
	rg.AddMetrics(m2)

	for i := 0; i < 100; i++ {
		m1.Inc()
		m2.Inc()
	}
	for i := 0; i < 100; i++ {
		m2.Inc()
	}

	assertCounter(t, 100, m1.Get())
	assertCounter(t, 200, m2.Get())
}

func TestTrackRegistry(t *testing.T) {
	rg, _ := metrics.NewTrackRegistry("testTrackRegistry", 10, time.Second, false)
	m1 := metrics.NewCounter("test swap metric")
	rg.AddMetrics(m1)

	for i := 0; i < 100; i++ {
		m1.Inc()
	}

	time.Sleep(time.Second + time.Millisecond*50)
	assertCounter(t, 0, m1.Get())

	sw := rg.GetSnapshots()
	metric, err := sw[0].GetMetricByName("test swap metric")
	if err != nil {
		t.Errorf("error on getting metric: %v", err)
	}
	assertCounter(t, 100, metric.Get())
}

func TestNewRegistry(t *testing.T) {
	rg, _ := metrics.NewRegistry("plain")
	m1 := metrics.NewCounter("metric")
	err := rg.AddMetrics(m1)
	if err != nil {
		t.Errorf("unable to register metric, %v", err)
	}

	for i := 0; i < 1000; i++ {
		m1.Inc()
	}
	assertCounter(t, 1000, m1.Get())

	m, err := rg.GetMetricByName("metric")
	if err != nil {
		t.Errorf("unable to get metric, %v", err)
	}
	assertCounter(t, 1000, m.Get())
}

func TestGauge(t *testing.T) {
	g := metrics.NewGauge("tgmetric")

	g.Add(42)
	g.Add(12.46)
	assertGauge(t, 54.46, g.Get())

	g.Sub(13.13)
	assertGauge(t, 41.33, g.Get())

	g.Set(0)
	assertGauge(t, 0, g.Get())

	reg, _ := metrics.NewTrackRegistry("gauge reg", 10, time.Second*5, true)
	err := reg.AddMetrics(g)

	if err != nil {
		t.Errorf("unbale to add gague into registry %v", err)
	}
}

func TestSnapshots(t *testing.T) {
	r, err := metrics.NewTrackRegistry("Test snaphots", 10, time.Second*5, false)
	tshould := time.Now().UTC().Add(time.Second * 5)
	if err != nil {
		t.Errorf("error on registry creation: %s", err)
	}
	g := metrics.NewGauge("testggauge")

	r.AddMetrics(g)

	for i := 0; i < 6; i++ {
		g.Add(10)
		time.Sleep(time.Second)
	}

	for _, s := range r.GetSnapshots() {
		m, err := s.GetMetricByName("testggauge")
		if err != nil {
			t.Errorf("unable to get metric from snapshot: %s", err)
		}
		assertGauge(t, 60, m.Get().(float64)+g.Get().(float64))

		if tshould.Round(time.Second) != s.GetTimestamp().Round(time.Second) {
			t.Errorf("snapshot time mismatch, expected %s, but got %s", tshould.Format("15:04:05"), s.GetTimestamp().Format("15:04:05"))
		}
	}
}

func assertGauge(t *testing.T, expected float64, actual interface{}) {
	if expected != actual.(float64) {
		t.Errorf("gauge mismatch, expected %f, but got %f", expected, actual)
	}
}

func assertCounter(t *testing.T, expected uint64, actual interface{}) {
	if expected != actual.(uint64) {
		t.Errorf("counter mismatch, expected %d, but got %d", expected, actual)
	}
}
