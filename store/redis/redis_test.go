package redis

import (
	"fmt"
	"testing"
	"time"

	"github.com/dvirsky/timedis/events"
	"github.com/stretchr/testify/assert"
)

func TestEncodeRecord(t *testing.T) {

	tm, _ := time.Parse("2006-Jan-02", "2012-Jul-09")
	r := events.Record{
		Time:  tm,
		Value: 1337,
	}

	assert.Equal(t, "000000004ffa1f00::1337", encodeRecord(r))
	fmt.Println(encodeRecord(r))
	r.Value = "foo"

	enc := encodeRecord(r)
	assert.Equal(t, `000000004ffa1f00::"foo"`, enc)

	r2, err := decodeRecord(enc)
	assert.NoError(t, err)
	assert.Equal(t, r.Value, r2.Value)
	assert.Equal(t, r.Time.Unix(), r2.Time.Unix())
}

func TestEncodeDecodeTime(t *testing.T) {

	tt := time.Now()
	ts := encodeTime(tt)

	tt2, err := decodeTime(ts)
	assert.NoError(t, err)

	assert.Equal(t, tt.Unix(), tt2.Unix())
}

func TestEncodeRange(t *testing.T) {

	fr, to := formatRange(time.Now(), time.Now().Add(time.Minute))
	//	fmt.Println(fr, to)
	assert.True(t, fr < to)

}

func TestPut(t *testing.T) {

	store := NewStore("localhost:6379")

	store.Put(
		events.Event{
			Key: "foo.bar",
			Record: events.Record{
				Value: 1337,
				Time:  time.Now(),
			},
		},
	)

}
