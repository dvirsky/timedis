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

func NewEvent(key string, t time.Time, val interface{}) Event {
	return Event{
		Key: key,
		Record: Record{
			Time:  t,
			Value: val,
		},
	}
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
