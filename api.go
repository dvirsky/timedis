package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/EverythingMe/vertex"
	"github.com/EverythingMe/vertex/middleware"
	"github.com/dvirsky/go-pylog/logging"
	"github.com/dvirsky/timedis/events"
	"github.com/dvirsky/timedis/sampler"
)

var engine *Engine

type EntryHandler struct {
	Key       string `schema:"key" maxlen:"1000" pattern:"[a-zA-Z_\.]+" required:"true" doc:"The key we are posting the event to" in:"query"`
	Value     string `schema:"value" maxlen:"1000" minlen:"1" required:"true" doc:"The value we are putting. Will be parsed as number or string"`
	Timestamp string `schema:"time" maxlen:"32" required:"false" doc:"The entry's timestamp in the form of "2006-01-02 15:04:05" (assuming gmt). If missing, the current timestamp will be used"`
}

const timeFormat = "2006-01-02 15:04:05"

func (h EntryHandler) Handle(w http.ResponseWriter, r *vertex.Request) (interface{}, error) {

	val, err := decodeValue(h.Value)
	if err != nil {
		return nil, err
	}

	tm := time.Now()
	if h.Timestamp != "" {
		if t, err := decodeTimestamp(h.Timestamp); err == nil {
			tm = t
		} else {
			logging.Warning("Error parsing time; %s", err)
		}
	}

	return "OK", engine.Store.Put(events.NewEvent(h.Key, tm, val))

}

type RangeHandler struct {
	Key  string `schema:"key" maxlen:"1000" pattern:"[a-zA-Z_\.]+" required:"true" doc:"The key we want data for" in:"query"`
	From string `schema:"from" maxlen:"32" required:"true" doc:"range start time, formatted as "2006-01-02 15:04:05" (assuming gmt)"`
	To   string `schema:"to" maxlen:"32" required:"false" doc:"range end time, formatted as "2006-01-02 15:04:05" (assuming gmt). If not present we default to now"`
}

func (h RangeHandler) Handle(w http.ResponseWriter, r *vertex.Request) (interface{}, error) {

	f, err := decodeTimestamp(h.From)
	if err != nil {
		return nil, err
	}
	t := time.Now()
	if h.To != "" {
		if t, err = decodeTimestamp(h.To); err != nil {
			return nil, err
		}
	}

	return engine.Store.Get(h.Key, f, t)
}

type SampleCounterHandler struct {
	Key   string  `schema:"key" maxlen:"1000" pattern:"[a-zA-Z_\.]+" required:"true" doc:"The key we want data for" in:"query"`
	Value float32 `schema:"value" required:"true"`
	Rate  float32 `schema:"rate" required:"false" default:"1.0" doc:"sample rate. defaults to 1.0"`
}

func (h SampleCounterHandler) Handle(w http.ResponseWriter, r *vertex.Request) (interface{}, error) {

	return "OK", engine.Sampler.Sample(h.Key, h.Value, h.Rate, sampler.SampleCounter)
}

type SampleTimerHandler SampleCounterHandler

func (h SampleTimerHandler) Handle(w http.ResponseWriter, r *vertex.Request) (interface{}, error) {

	return "OK", engine.Sampler.Sample(h.Key, h.Value, h.Rate, sampler.SampleTimer)
}

type SubscribeHandler struct {
	Key string `schema:"key" maxlen:"1000" pattern:"[a-zA-Z_\.]+" required:"true" doc:"The key we are subscribing to" in:"query"`
}

func (h SubscribeHandler) Handle(w http.ResponseWriter, r *vertex.Request) (interface{}, error) {

	ch, err := engine.Store.Subscribe(h.Key)
	if err != nil {
		return nil, err
	}

	flusher, _ := w.(http.Flusher)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	fmt.Fprintf(w, "retry: 500\n\n")
	flusher.Flush()

	for record := range ch {

		b, err := json.Marshal(record)
		if err == nil {
			_, err := fmt.Fprintf(w, "event: record\ndata: %s\n\n", string(b))

			if err == nil {
				flusher.Flush()
			} else {
				logging.Error("Could not send message to subscriber: %s, quitting", err)
				// TODO: close channel so we won't leak
				break
			}

		} else {
			logging.Error("Could not write message to json", err)
		}

	}

	return nil, vertex.Hijacked
}

func decodeTimestamp(ts string) (time.Time, error) {
	return time.Parse(timeFormat, ts)
}

func decodeValue(val string) (interface{}, error) {
	var err error

	if num, err := strconv.ParseInt(val, 10, 64); err == nil {
		return num, nil
	}

	if num, err := strconv.ParseUint(val, 10, 64); err == nil {
		return num, nil
	}

	if num, err := strconv.ParseFloat(val, 64); err == nil {
		return num, nil
	}

	if val == "true" {
		return true, nil
	}
	if val == "false" {
		return false, nil
	}

	val, err = strconv.Unquote(val)
	val_ := val
	if err != nil {
		logging.Error("Error unquoting value '%s': %s", val_, err)
		return nil, err
	}

	return val, nil
}

func init() {
	root := "/"
	vertex.Register("testung", func() *vertex.API {

		return &vertex.API{
			Name:          "testung",
			Version:       "1.0",
			Root:          root,
			Doc:           "Our fancy testung API",
			Title:         "TestungAPI",
			Middleware:    middleware.DefaultMiddleware,
			Renderer:      vertex.JSONRenderer{},
			AllowInsecure: vertex.Config.Server.AllowInsecure,
			SwaggerMiddleware: vertex.MiddlewareChain(
				middleware.NewCORS().Default(),
				middleware.NewIPRangeFilter().AllowPrivate(),
			),

			Routes: vertex.Routes{
				{
					Path:        "/entry/{key}",
					Description: "Post an entry into a key",
					Handler:     EntryHandler{},
					Methods:     vertex.POST | vertex.GET, // TODO: Remove GET
					Returns:     "OK",
				},
				{
					Path:        "/sample/counter/{key}",
					Description: "Post a counter sample",
					Handler:     SampleCounterHandler{},
					Methods:     vertex.POST | vertex.GET, // TODO: Remove GET
					Returns:     "OK",
				},
				{
					Path:        "/sample/timer/{key}",
					Description: "Post a timer sample",
					Handler:     SampleTimerHandler{},
					Methods:     vertex.POST | vertex.GET, // TODO: Remove GET
					Returns:     "OK",
				},
				{
					Path:        "/range/{key}",
					Description: "Get the values in a time range",
					Handler:     RangeHandler{},
					Methods:     vertex.GET,
					Returns:     events.Result{},
				},
				{
					Path:        "/subscribe/{key}",
					Description: "Subscribe to changes in a series",
					Handler:     SubscribeHandler{},
					Methods:     vertex.GET,
					Returns:     events.Result{},
				},

				{
					Path:        "/html/*filepath",
					Description: "Static",
					Handler:     vertex.StaticHandler(path.Join(root, "html"), http.Dir("./html")),
					Methods:     vertex.GET,
				},
			},
		}
	}, nil)
}
