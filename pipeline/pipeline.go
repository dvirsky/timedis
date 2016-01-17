package pipeline

import (
	"errors"
	"fmt"
	"time"

	"github.com/dvirsky/go-pylog/logging"
	"github.com/dvirsky/timedis/events"
	stor "github.com/dvirsky/timedis/store"
	"github.com/mitchellh/mapstructure"
)

var store stor.Store

func InitStore(s stor.Store) {
	store = s
}

type Source interface {
	Stream() (<-chan *events.Event, chan<- bool, error)
}

type SourceFactory func(params map[string]interface{}, upstream []Source) (Source, error)

func makeDownstream() (chan *events.Event, chan bool) {
	return make(chan *events.Event), make(chan bool)
}

type Filter struct {
	upstream Source
	MinValue float64 `mapstructure:"min"`
	MaxValue float64 `mapstructure:"max"`
}

func NewFilter(params map[string]interface{}, upstream []Source) (Source, error) {

	if len(upstream) != 1 {
		return nil, fmt.Errorf("Filter can have just 1 upstream, has %d", len(upstream))
	}

	ret := &Filter{}

	if err := decodeParams(params, ret); err != nil {
		return nil, err
	}
	ret.upstream = upstream[0]
	return ret, nil
}

func (f Filter) Stream() (<-chan *events.Event, chan<- bool, error) {

	events, stopch, err := f.upstream.Stream()
	if err != nil {
		return nil, nil, err
	}

	ret, stopret := makeDownstream()

	go func() {

		select {
		case ev, ok := <-events:
			if !ok {
				goto abort
			}
			if ev.Record.Value <= f.MaxValue && ev.Record.Value >= f.MinValue {
				ret <- ev
			}
		case <-stopret:
			goto abort

		}
	abort:
		close(ret)
		stopch <- true
		return
	}()

	return ret, stopret, nil
}

type MovingAverage struct {
	WindowSize int `mapstructure:"window"`
	upstream   Source
}

func decodeParams(params map[string]interface{}, value interface{}) error {
	return mapstructure.Decode(params, value)
}

func NewMovingAverage(params map[string]interface{}, upstream []Source) (Source, error) {
	if len(upstream) != 1 {
		return nil, errors.New("Filter can have just 1 upstream")
	}

	ret := &MovingAverage{}
	if err := decodeParams(params, ret); err != nil {
		return nil, err
	}
	ret.upstream = upstream[0]
	return ret, nil

}

func (f *MovingAverage) Stream() (<-chan *events.Event, chan<- bool, error) {

	events, stopch, err := f.upstream.Stream()
	if err != nil {
		return nil, nil, err
	}

	ret, stopret := makeDownstream()

	go func() {

		var average float64 = 0
		numSamples := 0
		for {
			select {
			case ev, ok := <-events:
				if !ok {
					goto abort
				}

				numSamples++
				if numSamples > f.WindowSize {
					average -= average / float64(f.WindowSize)
				}
				average += ev.Value / float64(f.WindowSize)
				if numSamples >= f.WindowSize {
					ev.Value = average
					ret <- ev
				}

			case <-stopret:
				logging.Warning("Got stop from downstream")
				goto abort

			}
		}
	abort:
		logging.Error("Aborting moving average")
		close(ret)
		stopch <- true
		return
	}()

	return ret, stopret, nil
}

type Faucet struct {
	Key  string `mapstructure:"key"`
	From int64  `mapstructure:"from"`
}

func NewFaucet(params map[string]interface{}, upstream []Source) (Source, error) {

	ret := &Faucet{}

	if err := decodeParams(params, ret); err != nil {
		return nil, err
	}

	if ret.Key == "" {
		return nil, errors.New("No key provided for faucet")
	}

	return ret, nil
}

func (f *Faucet) Stream() (<-chan *events.Event, chan<- bool, error) {

	from := time.Now().Add(time.Duration(f.From) * time.Second)

	results, err := store.Get(f.Key, from, time.Now())
	if err != nil {
		return nil, nil, err
	}

	evs, err := store.Subscribe(f.Key)
	if err != nil {
		return nil, nil, err
	}
	ret, stopret := makeDownstream()

	go func() {

		for _, rec := range results.Records {
			ret <- events.NewEvent(results.Key, rec.Time, rec.Value)
		}

		for res := range evs {
			for _, rec := range res.Records {

				ret <- events.NewEvent(res.Key, rec.Time, rec.Value)
			}
		}
	}()
	return ret, stopret, nil

}
