package main

import (
	"time"

	"github.com/EverythingMe/vertex"
	"github.com/dvirsky/go-pylog/logging"
	"github.com/dvirsky/timedis/events"
	"github.com/dvirsky/timedis/store/redis"
)

type Sampler interface {
	PutSample(...events.Sample) error
}
type Store interface {
	Put(...events.Event) error
	Get(key string, from, to time.Time) (events.Result, error)
	Subscribe(key string) (<-chan events.Result, error)
}

type Engine struct {
	Sampler Sampler
	Store   Store
}

func main() {

	engine = &Engine{
		Store: redis.NewStore("localhost:6379"),
	}

	vertex.ReadConfigs()

	logging.SetMinimalLevelByName(vertex.Config.Server.LoggingLevel)
	srv := vertex.NewServer(vertex.Config.Server.ListenAddr)
	srv.InitAPIs()
	if err := srv.Run(); err != nil {
		panic(err)
	}

}
