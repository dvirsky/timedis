package events

import (
	"encoding/json"
	"time"
)

type Record struct {
	Time  time.Time
	Value interface{}
}

func (r Record) MarshalJSON() ([]byte, error) {

	s := struct {
		Time  int64       `json:"time"`
		Value interface{} `json:"y"`
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
