package sampler

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {

	c := newCounter("foo")
	assert.NotNil(t, c)
	assert.NoError(t, c.Update(1, 1))
	assert.NoError(t, c.Update(2, 0.1))

	evs := c.Extract()
	assert.Len(t, evs, 1)

	assert.Equal(t, evs[0].Key, "foo")
	assert.Equal(t, evs[0].Value, float32(21))
	assert.Equal(t, evs[0].Time.Unix(), time.Now().Unix())
	fmt.Printf("%#v", c.Extract())
}

func TestSampler(t *testing.T) {
	s := NewSampler(time.Second, nil)

	assert.NotNil(t, s)

	assert.NoError(t, s.Sample("foo", 1, 1, SampleCounter))
	assert.NoError(t, s.Sample("foo", 1, 0.5, SampleCounter))
	assert.NoError(t, s.Sample("bar", 1, 0.2, SampleCounter))

	events := s.flush()
	assert.True(t, len(events) == 2)
	assert.True(t, len(s.samples) == 0)

	for _, ev := range events {
		if ev.Key != "foo" && ev.Key != "bar" {
			t.Error("Bad key", ev.Key)
		}
	}

	fmt.Println(events)

}
