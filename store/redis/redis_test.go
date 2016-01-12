package redis

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/dvirsky/timedis/events"
	"github.com/stretchr/testify/assert"
)

func TestEncodeRecord(t *testing.T) {

	tm, _ := time.Parse("2006-Jan-02", "2012-Jul-09")
	r := events.Record{
		Time:  tm,
		Value: 1,
	}

	assert.Equal(t, "000000004ffa1f00::1", encodeRecord(r))
	fmt.Println(decodeRecord(encodeRecord(r)))
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

func TestSubscribe(t *testing.T) {

	store := NewStore("localhost:6379")
	k := "foo.pbsb"

	sub, err := store.Subscribe(k)
	assert.NoError(t, err)
	var res events.Result
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		res = <-sub
		wg.Done()
	}()

	rec := events.Record{
		Value: "foo",
		Time:  time.Now(),
	}

	store.Put(
		events.Event{
			Key:    k,
			Record: rec,
		},
	)

	wg.Wait()
	assert.Equal(t, res.Key, k)
	assert.Len(t, res.Records, 1)
	assert.Equal(t, res.Records[0].Value, rec.Value)
	fmt.Println(res)

}

func TestGet(t *testing.T) {
	store := NewStore("localhost:6379")
	k := "test.key"
	tm, _ := time.Parse("2006-Jan-02", "2012-Jul-09")
	conn, _ := store.conn()
	conn.Do("DEL", store.dataKey(k))

	for i := 0; i < 10; i++ {
		store.Put(
			events.Event{
				Key: k,
				Record: events.Record{
					Value: i,
					Time:  tm.Add(time.Duration(i) * time.Second),
				},
			},
		)
	}

	res, err := store.Get(k, tm, tm.Add(5*time.Second))
	assert.NoError(t, err)
	assert.Len(t, res.Records, 6)
	assert.Equal(t, res.Key, k)
	for i, rec := range res.Records {
		assert.Equal(t, rec.Value.(int64), int64(i))
		assert.Equal(t, rec.Time.Unix(), tm.Add(time.Duration(i)*time.Second).Unix())
	}
}
