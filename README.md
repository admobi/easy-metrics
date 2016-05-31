# Easy-metrics [![GoDoc] (https://godoc.org/github.com/admobi/easy-metrics?status.svg)](https://godoc.org/github.com/admobi/easy-metrics) [![Go Report Card](https://goreportcard.com/badge/github.com/admobi/easy-metrics)](https://goreportcard.com/report/github.com/admobi/easy-metrics) [![Build Status](https://travis-ci.org/admobi/easy-metrics.svg?branch=master)](https://travis-ci.org/admobi/easy-metrics)
A Go library that provides easy to use, stand alone metrics and exposes it via HTTP.
It provides metrics snapshots on defined interval and stores it in a pool.
That can be used to track performance and other application indicators without necessity of export into external applications.

# Features
Metrics can be stored in a `TrackRegistry` that provides snapshots over defined interval and stores it in a pool. Snapshots like other metrics available via HTTP.
So you don't need to collect metrics in some external applications like statsd, graphana, elk etc.

# Installation
Just go get:
```
go get github.com/admobi/easy-metrics
```

Or to update:
``` 
go get -u github.com/admobi/easy-metrics
```

# Usage
At the core of metrics is two subjects, *metric* which stores a single numerical value
and *registry* which stores pool of metrics

Add import to project:
```go
import "github.com/admobi/easy-metrics"
```

Create and update metrics:
```go
// Create metric
c := mertics.NewCounter("requests")
// Create registry
r := metrics.NewRegistry("Statistics")
// Register metric
r.AddMetrics(c)

// Change metric. Increase by 1 
c.Inc()
// or same
c.Add(1)
```

You can add several metrics into registry:
```go
// Create metric
c := mertics.NewCounter("requests")
g := mertics.NewGauge("rates")
// Create registry
r, err := metrics.NewRegistry("Statistics")
// Register metric
r.AddMetrics(c, g)
```

## Snapshots
```go
r := metrics.NewTrackRegistry("Stat", 30, time.Second, false)
c := mertics.NewCounter("requests")
r.AddMetrics(c)
```

TrackRegistry will take metric snapshot every second and stores 30 last results

All operations are thread safe