package modules

import (
	"fmt"
	"regexp"
	"testing"
)

func Test_getDiskStats(t *testing.T) {
	r, _ := regexp.Compile(DiskNameRegexp)
	t.Run("", func(t *testing.T) {
		got, _ := getDiskStats(r)

		// just check if we got some values
		if len(got) == 0 {
			t.Errorf("Did not get any statistics")
		}

		// we are expecting to see statistics for the SD card, i.e. mmcblk
		if _, ok := got["mmcblk0"]; !ok {
			t.Errorf("Did not find mmcblk0 key")
		}

		// if we have anything else, check that we have some sensible numbers
		for k, elem := range got {
			if elem.ReadIOs == 0 {
				t.Errorf("%s did not read any ReadIOs, got %v", k, elem.ReadIOs)
			}
			if elem.ReadMerges == 0 {
				t.Errorf("%s did not read any ReadMerges, got %v", k, elem.ReadMerges)
			}
		}
	})

}

func Test_formatDiskStatsLine(t *testing.T) {
	var tests = []struct {
		inputLine string
		want      string
	}{
		{"   1       0 ram0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0",
			"1 0 ram0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0"},
		{"   1      10 ram10 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0",
			"1 10 ram10 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0"},
		{"   8       0 sda 52817 160284 5969575 1616869 14792321 2261382 1286890664 204844813 0 93397930 206525131 0 0 0 0 157356 63448",
			"8 0 sda 52817 160284 5969575 1616869 14792321 2261382 1286890664 204844813 0 93397930 206525131 0 0 0 0 157356 63448"},
		{" 179       0 mmcblk0 287277 162611 16233367 1243501 473290 864375 17033130 13592893 0 3950880 14836394 0 0 0 0 0 0",
			"179 0 mmcblk0 287277 162611 16233367 1243501 473290 864375 17033130 13592893 0 3950880 14836394 0 0 0 0 0 0"},
	}

	for i, tt := range tests {
		testname := fmt.Sprintf("%d", i)
		t.Run(testname, func(t *testing.T) {
			got := formatDiskStatsLine(tt.inputLine)
			if got != tt.want {
				t.Errorf("Got %s, want %s", got, tt.want)
			}
		})
	}
}

func Test_getDiskStatsFromLine(t *testing.T) {
	var tests = []struct {
		sampleLine string
		wantKey    string
		wantStats  DiskStats
	}{
		{"8 0 sda 52817 160284 5969575 1616869 14792321 2261382 1286890664 204844813 0 93397930 206525131 0 0 0 0 157356 63448",
			"sda",
			DiskStats{
				ReadIOs:       52817,
				ReadMerges:    160284,
				FlushSuccess:  157356,
				FlushingTicks: 63448}},
		{"179 0 mmcblk0 287277 162611 16233367 1243501 473290 864375 17033130 13592893 0 3950880 14836394 0 0 0 0 0 0",
			"mmcblk0",
			DiskStats{
				ReadIOs:       287277,
				ReadMerges:    162611,
				FlushSuccess:  0,
				FlushingTicks: 0}},
	}

	for i, tt := range tests {
		testname := fmt.Sprintf("%d", i)
		t.Run(testname, func(t *testing.T) {
			gotStats := getDiskStatsFromLine(tt.sampleLine)

			if gotStats.ReadIOs != tt.wantStats.ReadIOs {
				t.Errorf("ReadIOs for %s, got %v, want %v",
					gotStats.DevName, gotStats.ReadIOs, tt.wantStats.ReadIOs)
			}
			if gotStats.ReadMerges != tt.wantStats.ReadMerges {
				t.Errorf("ReadMerges for %s, got %v, want %v",
					gotStats.DevName, gotStats.ReadMerges, tt.wantStats.ReadMerges)
			}
			if gotStats.FlushSuccess != tt.wantStats.FlushSuccess {
				t.Errorf("FlushSuccess for %s, got %v, want %v",
					gotStats.DevName, gotStats.FlushSuccess, tt.wantStats.FlushSuccess)
			}
			if gotStats.FlushingTicks != tt.wantStats.FlushingTicks {
				t.Errorf("FlushingTicks for %s, got %v, want %v",
					gotStats.DevName, gotStats.FlushingTicks, tt.wantStats.FlushingTicks)
			}

		})
	}
}
