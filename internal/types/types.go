package types

import (
	"encoding/binary"
	"math"
	"strconv"
)

type (
	Gauge   float64
	Counter int64
)

func (g Gauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}

func GaugeToBytes(g Gauge) []byte {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], math.Float64bits(float64(g)))
	return buf[:]
}

func BytesToGauge(b []byte) Gauge {
	bits := binary.LittleEndian.Uint64(b)
	return Gauge(math.Float64frombits(bits))
}

func ParseGauge(in string) (Gauge, error) {
	value, err := strconv.ParseFloat(in, 64)
	return Gauge(value), err
}

func (c Counter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

func CounterToBytes(c Counter) []byte {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(c))
	return buf[:]
}

func BytesToCounter(b []byte) Counter {
	res := binary.LittleEndian.Uint64(b)
	return Counter(res)
}

func ParseCounter(in string) (Counter, error) {
	value, err := strconv.ParseInt(in, 10, 64)
	return Counter(value), err
}
