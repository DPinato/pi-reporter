package modules

import (
	"fmt"
	"testing"
)

func Test_getMemoryStats(t *testing.T) {
	t.Run("Read memory statistics", func(t *testing.T) {
		got, err := getMemoryStats()

		// check for error
		if err != nil {
			t.Errorf("Got error, %v\n", err)
		}

		// check for impossible values
		for key, elem := range got {
			if elem < 0 {
				t.Errorf("Retrieved invalid memory stat value, %s %d\n", key, elem)
			}
		}
	})
}

func Test_getMemoryStatFromLine(t *testing.T) {
	var tests = []struct {
		sampleLine string
		wantField  string
		wantValue  int
	}{
		{"MemTotal:         992964 kB", "MemTotal", 992964},
		{"Writeback:             0 kB", "Writeback", 0},
		{"CmaFree:            5952 kB", "CmaFree", 5952},
		{"TestBadLine", "", -1},
		{"TestBadLine2:", "TestBadLine2", -1},
		{"TestBadLine3:         ", "TestBadLine3", -1},
	}

	for i, tt := range tests {
		testname := fmt.Sprintf("%d", i)
		t.Run(testname, func(t *testing.T) {
			gotField, gotValue := getMemoryStatFromLine(tt.sampleLine)
			if gotField != tt.wantField {
				t.Errorf("Got field %s, want %s", gotField, tt.wantField)
			}
			if gotValue != tt.wantValue {
				t.Errorf("Got field %d, want %d", gotValue, tt.wantValue)
			}
		})
	}
}
