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

type Result struct {
	Records []Record
	Key     string
}
