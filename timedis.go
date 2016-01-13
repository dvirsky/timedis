package main

import (
	"time"

	"github.com/EverythingMe/vertex"
	"github.com/dvirsky/go-pylog/logging"
	"github.com/dvirsky/timedis/sampler"
	"github.com/dvirsky/timedis/store"
	"github.com/dvirsky/timedis/store/redis"
)

type Engine struct {
	Sampler *sampler.Sampler
	Store   store.Store
}

func main() {

	store := redis.NewStore("localhost:6379")
	sampler := sampler.NewSampler(time.Second, store)
	engine = &Engine{
		Store:   store,
		Sampler: sampler,
	}

	sampler.Run()

	vertex.ReadConfigs()

	logging.SetMinimalLevelByName(vertex.Config.Server.LoggingLevel)
	srv := vertex.NewServer(vertex.Config.Server.ListenAddr)
	srv.InitAPIs()
	if err := srv.Run(); err != nil {
		panic(err)
	}

}
