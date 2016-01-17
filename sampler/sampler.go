package sampler

import (
	"fmt"
	"sync"
	"time"

	"github.com/dvirsky/go-pylog/logging"
	"github.com/dvirsky/timedis/events"
	"github.com/dvirsky/timedis/store"
)

type SampleType int

const (
	SampleCounter SampleType = iota
	SampleTimer
	SampleHistogram
)

type sample interface {
	Update(value, rate float64) error
	Extract() []*events.Event
}

type baseSample struct {
	Key  string
	lock sync.Mutex
}

type counter struct {
	baseSample
	Duration time.Duration
	Value    float64
}

func (c *counter) Update(value, rate float64) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.Value += value / rate

	return nil
}

func (c *counter) Extract() []*events.Event {

	return []*events.Event{events.NewEvent(c.Key, time.Now(), c.Value)}

}

type timer struct {
	baseSample
	Value      float64
	NumSamples float64
}

func (c *timer) Update(value, rate float64) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.Value += value / rate
	c.NumSamples += 1 / rate
	return nil
}

func (c *timer) Extract() []*events.Event {

	if c.NumSamples == 0 {
		return nil
	}
	return []*events.Event{events.NewEvent(c.Key, time.Now(), c.Value/c.NumSamples)}

}

type Sampler struct {
	lock    sync.Mutex
	samples map[string]sample
	tick    time.Duration
	store   store.Store
}

func NewSampler(tick time.Duration, st store.Store) *Sampler {

	return &Sampler{
		lock:    sync.Mutex{},
		samples: make(map[string]sample),
		tick:    tick,
		store:   st,
	}
}

func (s *Sampler) Run() {

	go func() {

		for range time.Tick(s.tick) {

			go func() {

				events := s.flush()
				if events != nil && len(events) > 0 {
					logging.Info("Flushing %d events", events)
					s.store.Put(events...)
				}
			}()
		}

	}()

}

func (s *Sampler) flush() []*events.Event {

	s.lock.Lock()
	samples := s.samples
	s.samples = make(map[string]sample)
	s.lock.Unlock()

	ret := make([]*events.Event, 0, len(samples))
	for _, sm := range samples {
		ret = append(ret, sm.Extract()...)
	}

	return ret

}

func (s *Sampler) get(key string, t SampleType) (sample, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if sm, found := s.samples[key]; found {
		return sm, nil
	}

	var ret sample
	switch t {
	case SampleCounter:
		ret = newCounter(key)
	case SampleTimer:
		ret = newTimer(key)
	default:
		return nil, fmt.Errorf("Unsupported sample type: %v", t)
	}
	s.samples[key] = ret
	return ret, nil
}

func (s *Sampler) Sample(key string, value, rate float64, t SampleType) error {

	smp, err := s.get(key, t)
	if err != nil {
		return err
	}

	return smp.Update(value, rate)

}

func newCounter(key string) *counter {

	c := &counter{
		baseSample: baseSample{
			Key:  key,
			lock: sync.Mutex{},
		},
		Duration: time.Second,
	}

	return c
}

func newTimer(key string) *timer {

	c := &timer{
		baseSample: baseSample{
			Key:  key,
			lock: sync.Mutex{},
		},
	}

	return c
}
