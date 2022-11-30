package metrics

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

/*
1. Create metrics packet--> packet with identifier to track latency, errors.

*/

type Metric struct {
	LastRecordedLatency uint64
	ConnectionStatus    bool
	TrafficSent         uint64
	TrafficRecieved     uint64
}

var MetricsMapLock = &sync.Mutex{}

var MetricsMap = make(map[string]Metric)

func init() {
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			PrintMetrics()
		}
	}()
}

func PrintMetrics() {

	data, err := json.MarshalIndent(MetricsMap, "", " ")
	if err != nil {
		return
	}
	os.WriteFile("/tmp/metrics.json", data, 0755)

}