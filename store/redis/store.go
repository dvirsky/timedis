package redis

import (
	"errors"
	"fmt"
	"time"

	"github.com/dvirsky/go-pylog/logging"
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

func (s *Store) conn() (redis.Conn, error) {

	conn := s.pool.Get()
	if conn == nil {
		return nil, errors.New("Could not connect")
	}
	return conn, nil
}

func (s *Store) Put(evs ...events.Event) error {

	conn, err := s.conn()
	if err != nil {
		return err
	}
	defer conn.Close()

	for _, ev := range evs {

		enc := encodeRecord(ev.Record)
		conn.Send("ZADD", s.dataKey(ev.Key), 0, enc)
		conn.Send("PUBLISH", s.pubsubKey(ev.Key), enc)

	}

	return conn.Flush()
	return nil
}

func (s *Store) dataKey(key string) string {
	return fmt.Sprintf("d::%s", key)
}
func (s *Store) pubsubKey(key string) string {
	return fmt.Sprintf("ps::%s", key)
}

func (s *Store) Get(key string, from, to time.Time) (events.Result, error) {
	conn, err := s.conn()
	if err != nil {
		return events.Result{}, err
	}
	defer conn.Close()

	f, t := formatRange(from, to)
	values, err := redis.Strings(conn.Do("ZRANGEBYLEX", s.dataKey(key), f, t))
	if err != nil {
		return events.Result{}, err
	}

	res := events.Result{
		Key:     key,
		Records: make([]events.Record, 0, len(values)),
	}

	for _, encoded := range values {

		rec, err := decodeRecord(encoded)
		if err != nil {
			logging.Error("Error decoding record: %s", err)
			continue
		}

		res.Records = append(res.Records, rec)

	}

	return res, nil

}

func (s *Store) pushUpdate(key, encodedRecord string) {

	conn, err := s.conn()
	if err != nil {
		logging.Warning("Could not get connection for pushing: %s", err)
		return
	}
	defer conn.Close()

	if _, err := conn.Do("PUBLISH", s.pubsubKey(key), encodedRecord); err != nil {
		logging.Error("Could not send pubsub: %s", err)
	}

}
func (s *Store) Subscribe(key string) (<-chan events.Result, error) {

	conn, err := s.conn()
	if err != nil {
		return nil, err
	}

	ch := make(chan events.Result)

	go func() {
		for {

			if conn == nil {
				if conn, err = s.conn(); err != nil {
					logging.Error("Error connecting to pubsub: %s", err)
				}
			} else {

				psc := redis.PubSubConn{conn}
				err := psc.Subscribe(s.pubsubKey(key))
				if err != nil {
					logging.Error("Could not subscribe: %s", err)
					time.Sleep(100 * time.Microsecond)
					conn.Close()
					conn = nil
					continue
				}

				for err == nil {
					switch v := psc.Receive().(type) {
					case redis.Message:
						logging.Info("Got an update for %s: %s", key, string(v.Data))

						if rec, err := decodeRecord(string(v.Data)); err != nil {
							logging.Warning("Could not decode pubsub message! %s", err)
							continue
						} else {
							ch <- events.Result{
								Key:     key,
								Records: []events.Record{rec},
							}
						}

					case redis.Subscription:
						logging.Debug("Subscription %s: %s %d\n", v.Channel, v.Kind, v.Count)
					case error:
						logging.Error("Error reading pubsub: %s", v)
						conn.Close()
						conn = nil
						err = v
						break
					}

				}
			}

			// sleep before retrying
			time.Sleep(100 * time.Millisecond)
		}

	}()

	return ch, nil

}
