package types

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
)

type (
	Gauge   float64
	Counter int64
)

// String returns Gauge as string.
func (g Gauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}

// GaugeToBytes returns Gauge as LE byte slice.
func GaugeToBytes(g Gauge) []byte {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], math.Float64bits(float64(g)))
	return buf[:]
}

// BytesToGauge converts LE byte slice to Gauge.
func BytesToGauge(b []byte) Gauge {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic: ", r)
		}
	}()
	bits := binary.LittleEndian.Uint64(b)
	return Gauge(math.Float64frombits(bits))
}

// ParseGauge returns Gauge from string if it parsed without error else it returns zero value with error.
func ParseGauge(in string) (Gauge, error) {
	value, err := strconv.ParseFloat(in, 64)
	if err == nil {
		return Gauge(value), nil
	}
	return Gauge(0), err
}

// String returns Counter as string.
func (c Counter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

// CounterToBytes returns Counter as LE byte slice.
func CounterToBytes(c Counter) []byte {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(c))
	return buf[:]
}

// BytesToCounter converts LE byte slice to Counter.
func BytesToCounter(b []byte) Counter {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic: ", r)
		}
	}()
	res := binary.LittleEndian.Uint64(b)
	return Counter(res)
}

// ParseCounter returns Counter from string if it parsed without error else it returns zero value with error.
func ParseCounter(in string) (Counter, error) {
	value, err := strconv.ParseInt(in, 10, 64)
	return Counter(value), err
}
