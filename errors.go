package metrics

// ErrEmptyMetricName error type on empty metric name.
type ErrEmptyMetricName struct{}

func (e ErrEmptyMetricName) Error() string {
	return "metric name is empty"
}

// ErrEmptyRegistryName error type on empty registry name.
type ErrEmptyRegistryName struct{}

func (e ErrEmptyRegistryName) Error() string {
	return "registry name is empty"
}

// ErrRegistryUnknown error type on provided registry name is unknown.
type ErrRegistryUnknown string

func (e ErrRegistryUnknown) Error() string {
	return "undefined registry name " + string(e)
}

// ErrRegistryExists error type - registry with provided is exists.
type ErrRegistryExists string

func (e ErrRegistryExists) Error() string {
	return "registry with given name exists: " + string(e)
}

// ErrMetricUnknown error type on provided metric name is unknown.
type ErrMetricUnknown string

func (e ErrMetricUnknown) Error() string {
	return "undefined metric name " + string(e)
}

// ErrMetricExists error type - metric with provided name exists.
type ErrMetricExists string

func (e ErrMetricExists) Error() string {
	return "metric with given name exists: " + string(e)
}
