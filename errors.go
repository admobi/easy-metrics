package easy-metrics

import "errors"

var (
	// ErrEmptyMetricName error on empty metric name
	ErrEmptyMetricName = errors.New("metric name is empty")

	// ErrEmptyRegistryName error on empty registry name
	ErrEmptyRegistryName = errors.New("registry name is empty")
)

// ErrRegistryUnknown error type on provided registry name is unknown
type ErrRegistryUnknown string

func (e ErrRegistryUnknown) Error() string {
	return "undefined registry name " + string(e)
}

// ErrRegistryExists error type - registry with provided is exists
type ErrRegistryExists string

func (e ErrRegistryExists) Error() string {
	return "registry with given name exists: " + string(e)
}

// ErrMetricUnknown error type on provided metric name is unknown
type ErrMetricUnknown string

func (e ErrMetricUnknown) Error() string {
	return "undefined metric name " + string(e)
}

// ErrMetricExists error type - metric with provided is exists
type ErrMetricExists string

func (e ErrMetricExists) Error() string {
	return "metric with given name exists: " + string(e)
}
