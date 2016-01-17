package events

import (
	"encoding/json"
	"time"
)

type Record struct {
	Time  time.Time
	Value float64
}

func (r Record) MarshalJSON() ([]byte, error) {

	s := struct {
		Time  int64   `json:"time"`
		Value float64 `json:"y"`
	}{
		Time:  r.Time.Unix(),
		Value: r.Value,
	}

	return json.Marshal(s)

}

type Event struct {
	Record
	Key string
}

func NewEvent(key string, t time.Time, val float64) *Event {
	return &Event{
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
