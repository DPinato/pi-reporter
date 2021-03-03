package modules

import (
	"fmt"
	"log"
	"testing"

	"github.com/dpinato/pi-reporter/helper"
)

func Test_getWiredNetworkUsageSample(t *testing.T) {
	ifName := helper.PINetIfaces[0]
	testname := fmt.Sprintf("%s", ifName)
	t.Run(testname, func(t *testing.T) {
		stat, err := getNetworkIfStatistics(ifName)

		// check for error
		if err != nil {
			t.Errorf("Got error, %v\n", err)
		}

		// check that we got all the expected values
		if len(stat.Statistics) != len(NetStatsList) {
			t.Errorf("Did not retrieve %d statistics, only %d\n", len(NetStatsList), len(stat.Statistics))
		}
	})
}

// benchmarks
func benchmarkGetNetworkIfStatistics(ifName string, b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := getNetworkIfStatistics(ifName)
		if err != nil {
			log.Printf("benchmark_getNetworkIfStatistics failed, %v\n", err)
			b.Failed()
		}
	}
}

func Benchmark_GetWiredStats(b *testing.B) { benchmarkGetNetworkIfStatistics(helper.PINetIfaces[0], b) }

func Benchmark_GetWirelessStats(b *testing.B) {
	benchmarkGetNetworkIfStatistics(helper.PINetIfaces[1], b)
}
