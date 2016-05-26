# Easy-metrics
A Go library that provides easy to use, stand by metrics and exposes it via HTTP.
It can create snapshots each defined duration and store in a pool.
That useful for fast monitoring of current performance and other application indicators without necessity of export into external applications.

# Features
Metrics can be stored in a `TrackRegistry` that will make snapshots over defined time and store it in pool. Shapshots like other metrics available via HTTP.
So you don't need to collect it in some external applications like statsd, graphana, elk etc.

# Installation
Just go get:
```
go get github.com/adbluehub/easy-metrics
```

Or to update:
``` 
go get -u github.com/adbluehub/easy-metrics
```

# Usage
At the core of metrics the two main subject, *metric* which stores a single numerical value
and *registry* which stores pool of metrics

Add import to project:
```go
import (
	metrics "github.com/adbluehub/easy-metrics"
)
```

Create and update metrics:
```go
// Create metric
c := mertics.NewCounter("requests")
// Create registry
r := metrics.NewRegistry("Statistics")
// Register metric
r.AddMetric(c)

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
r.AddMetric(c, g)
```

All operations are thred safe