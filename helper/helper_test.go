package helper

import (
	"fmt"
	"log"
	"regexp"
	"testing"
)

func Test_GetPIName(t *testing.T) {
	regexpStr := "pi-[0-9,a-f]{12}"
	r, err := regexp.Compile(regexpStr)
	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	for _, elem := range PINetIfaces {
		testname := fmt.Sprintf("match_regexp %s", elem)
		t.Run(testname, func(t *testing.T) {
			got, err := GetPIName(elem)
			if err != nil {
				log.Println(err)
				t.FailNow()
			}

			if !r.MatchString(got) {
				t.Errorf("Did not match regexp, got %s\n", got)
			}
		})
	}

}
