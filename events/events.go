package events

import "time"

type Record struct {
	Time  time.Time
	Value interface{}
}

type Event struct {
	Record
	Key string
}

type SampleType int

const (
	SampleSet SampleType = iota
	SampleIncrement
	SampleAverage
)

type Sample struct {
	Event
	Value      float64
	Type       SampleType
	SampleRate float64
}

type Result struct {
	Records []Record
	Key     string
}
