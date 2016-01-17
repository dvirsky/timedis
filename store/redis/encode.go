package redis

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dvirsky/timedis/events"
)

func encodeTime(t time.Time) string {
	tv := t.Unix()
	ret := make([]byte, 4)
	binary.BigEndian.PutUint32(ret, uint32(tv))
	return hex.EncodeToString(ret)
}

func decodeTime(ts string) (time.Time, error) {
	dat, err := hex.DecodeString(ts)
	if err != nil {
		return time.Now(), err
	}

	tv := binary.BigEndian.Uint32(dat)
	return time.Unix(int64(tv), 0), nil
}

func decodeValue(val string) (float64, error) {

	if num, err := strconv.ParseInt(val, 10, 64); err == nil {
		return float64(num), nil
	}

	if num, err := strconv.ParseFloat(val, 64); err == nil {
		return num, nil
	} else {
		return 0, err
	}
}

func formatRange(from, to time.Time) (string, string) {

	return fmt.Sprintf("[%s::", encodeTime(from)), fmt.Sprintf("[%s::\xff", encodeTime(to))
}

func decodeRecord(entry string) (rec events.Record, err error) {

	parts := strings.Split(entry, "::")
	if len(parts) != 2 {
		err = errors.New("invalid record: " + entry)
		return
	}

	if rec.Time, err = decodeTime(parts[0]); err != nil {
		return
	}

	if rec.Value, err = decodeValue(parts[1]); err != nil {
		return
	}
	return
}

func encodeRecord(r events.Record) string {
	return fmt.Sprintf("%s::%#v", encodeTime(r.Time), r.Value)
}
