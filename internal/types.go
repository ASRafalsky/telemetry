package internal

import (
	"strconv"
)

type (
	Gauge   float64
	Counter int64
)

func (g Gauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}

func ParseGauge(in string) (Gauge, error) {
	value, err := strconv.ParseFloat(in, 64)
	return Gauge(value), err
}

func (c Counter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

func ParseCounter(in string) (Counter, error) {
	value, err := strconv.ParseInt(in, 10, 64)
	return Counter(value), err
}
