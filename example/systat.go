package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/akhenakh/statgo"
	"github.com/dvirsky/go-pylog/logging"
)

func putStat(key string, val interface{}) error {

	res, err := http.Get(fmt.Sprintf("http://localhost:9944/entry/%s?value=%v", url.QueryEscape(key), val))
	if err != nil {
		logging.Error("Error sending stat: %s", err)
		return err
	}
	res.Body.Close()
	return nil
}

func sampleStats() {

	for range time.Tick(time.Second) {
		stat := statgo.NewStat()

		putStat("sys.cpu.user", stat.CPUStats().User)
		putStat("sys.cpu.idle", stat.CPUStats().Idle)
		putStat("sys.cpu.kernel", stat.CPUStats().Kernel)
		putStat("sys.cpu.load", stat.CPUStats().LoadMin1)
		putStat("sys.mem.used", stat.MemStats().Used)
		putStat("sys.mem.free", stat.MemStats().Free)
		putStat("sys.mem.cached", stat.MemStats().Cache)
		putStat("sys.mem.total", stat.MemStats().Total)
	}
}

func main() {
	sampleStats()
}
