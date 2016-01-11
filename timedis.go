package main

import (
	"fmt"
	"time"

	"github.com/dvirsky/timedis/events"
)

type Sampler interface {
	PutSample(...events.Sample) error
}
type Store interface {
	Put(...events.Event) error
	Get(key string, from, to time.Time) ([]events.Result, error)
	Subscribe(key string) (chan<- events.Result, error)
}

type Engine struct {
	Sampler Sampler
	Store   Store
}

func main() {
	fmt.Println("Hello world!")
}
