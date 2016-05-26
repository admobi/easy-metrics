package metric_test

import (
	"github.com/adbluehub/DSP/backend/metric"
	"testing"
	"time"
)

func BenchmarkCounter(b *testing.B) {
	m1 := metric.NewCounter("Total_connections")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m1.Inc()
	}
}

func BenchmarkGauge(b *testing.B) {
	m1 := metric.NewGauge("Total connections")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m1.Add(1e-13)
	}
}

func TestNewTrackRegistry(t *testing.T) {
	_, err := metric.NewTrackRegistry("newswap", 10, time.Second*100, true)
	if err != nil {
		t.Errorf("unable to create swap registry, %v", err)
	}

	_, err = metric.NewTrackRegistry("", 10, time.Second*100, false)

	if err == nil {
		t.Error("empty registry name, should be error but got nil")
	}

	_, err = metric.NewTrackRegistry("newswap", 10, time.Second*200, false)

	if err == nil {
		t.Error("existing registry name, should be error but got nil")
	}

	switch err.(type) {
	case metric.ErrRegistryExists:
	default:
		t.Errorf("undefined error %v", err)
	}
}

func TestRegisterMetric(t *testing.T) {
	rg, _ := metric.NewTrackRegistry("newcounter", 10, time.Second*100, false)
	m1 := metric.NewCounter("name1")
	err := rg.AddMetric(m1)
	if err != nil {
		t.Errorf("unable to register metric, %v", err)
	}
	m2 := metric.NewCounter("name1")
	err = rg.AddMetric(m2)
	if err == nil {
		t.Error("existing metric name, should be error on register but got nil")
	}

	switch err.(type) {
	case metric.ErrMetricExists:
	default:
		t.Errorf("undefined error %v", err)
	}
}

func TestDefaultCounter(t *testing.T) {
	rg, _ := metric.NewTrackRegistry("testcounters", 10, time.Second*100, false)
	m1 := metric.NewCounter("Total connections")
	rg.AddMetric(m1)
	m2 := metric.NewCounter("Total connections 2")
	rg.AddMetric(m2)

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
	rg, _ := metric.NewTrackRegistry("testTrackRegistry", 10, time.Second, false)
	m1 := metric.NewCounter("test swap metric")
	rg.AddMetric(m1)

	for i := 0; i < 100; i++ {
		m1.Inc()
	}

	time.Sleep(time.Second + time.Millisecond*50)
	assertCounter(t, 0, m1.Get())

	sw := rg.GetSnapshots()
	assertCounter(t, 100, sw[0].GetMetric("test swap metric").Get())
}

func TestDumpRegistry(t *testing.T) {
	rg, _ := metric.NewTrackRegistry("dump", 10, time.Millisecond*100, true)
	m1 := metric.NewCounter("dump")
	rg.AddMetric(m1)
	m2 := metric.NewCounter("dump2")
	rg.AddMetric(m2)

	for i := 0; i < 1000; i++ {
		m1.Inc()
		m2.Inc()
	}

	str, err := metric.DumpRegistry("dump")
	if err != nil {
		t.Errorf("unable to dump registry, %v", err)
	}
	if len(str) <= 0 {
		t.Errorf("no data in registry dump")
	}

	_, err = metric.DumpRegistry("noexists")
	if err == nil {
		t.Error("have wrong registry name, should be error but got nil")
	}
}

func TestNewRegistry(t *testing.T) {
	rg, _ := metric.NewRegistry("plain")
	m1 := metric.NewCounter("metric")
	err := rg.AddMetric(m1)
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
	g := metric.NewGauge("tgmetric")

	g.Add(42)
	g.Add(12.46)
	assertGauge(t, 54.46, g.Get())

	g.Sub(13.13)
	assertGauge(t, 41.33, g.Get())

	g.Set(0)
	assertGauge(t, 0, g.Get())

	reg, _ := metric.NewTrackRegistry("gauge reg", 10, time.Millisecond, false)
	err := reg.AddMetric(g)

	time.Sleep(time.Millisecond * 5)
	if err != nil {
		t.Errorf("unbale to add gague into registry %v", err)
	}
}

func TestSnapshots(t *testing.T) {
	t.Error("Should be implemented")
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
