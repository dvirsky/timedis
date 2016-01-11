package redis

import (
	"time"

	"github.com/dvirsky/timedis/events"
	"github.com/garyburd/redigo/redis"
)

type Store struct {
	pool *redis.Pool
}

func NewStore(addr string) *Store {
	return &Store{
		pool: &redis.Pool{
			MaxIdle:     10,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {

				c, err := redis.Dial("tcp", addr)

				if err != nil {
					return nil, err
				}
				return c, err
			},
			TestOnBorrow: func(c redis.Conn, pooledTime time.Time) error {

				// for connections that were idle for over a second, let's make sure they can still talk to redis before doing anything with them
				if time.Since(pooledTime) > time.Second {
					_, err := c.Do("PING")
					return err
				}
				return nil
			},
		},
	}
}

func (s *Store) Put(evs ...events.Event) error {

	conn := s.pool.Get()

	for _, ev := range evs {

		enc := encodeRecord(ev.Record)
		conn.Send("ZADD", ev.Key, 0, enc)

	}

	return conn.Flush()
	return nil
}

func (s *Store) Get(key string, from, to time.Time) ([]events.Result, error) {
	return nil, nil
}
func (s *Store) Subscribe(key string) (chan<- events.Result, error) {
	return nil, nil
}
