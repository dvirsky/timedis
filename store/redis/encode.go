package redis

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dvirsky/go-pylog/logging"
	"github.com/dvirsky/timedis/events"
)

func encodeTime(t time.Time) string {
	tv := t.Unix()
	ret := make([]byte, binary.Size(tv))
	binary.BigEndian.PutUint64(ret, uint64(tv))
	return hex.EncodeToString(ret)
}

func decodeTime(ts string) (time.Time, error) {
	dat, err := hex.DecodeString(ts)
	if err != nil {
		return time.Now(), err
	}

	tv := binary.BigEndian.Uint64(dat)
	return time.Unix(int64(tv), 0), nil
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
