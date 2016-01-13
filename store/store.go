package store

import (
	"time"

	"github.com/dvirsky/timedis/events"
)

type Store interface {
	Put(...events.Event) error
	Get(key string, from, to time.Time) (events.Result, error)
	Subscribe(key string) (<-chan events.Result, error)
}
