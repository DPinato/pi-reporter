package modules

import (
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/dpinato/pi-reporter/helper"
	client "github.com/influxdata/influxdb1-client/v2"
)

const DefaultMemoryReportTime = 30 * time.Second
const MemoryStatsPath = "/proc/meminfo"
const MemoryMeasurementsName = "memory_stats"

func ReportMemoryStats(dbName string, c client.Client) error {
	myName, _ := helper.GetPIName(helper.PINetIfaces[0])
	log.Printf("ReportMemoryStats() is starting, %s\n", myName)

	ticker := time.NewTicker(DefaultMemoryReportTime)
	for {
		select {
		case t := <-ticker.C:
			stat, err := getMemoryStats()
			if err != nil {
				log.Println(err)
				continue
			}

			err = reportMemoryStatsToInflux(dbName, myName, stat, t, c)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func getMemoryStats() (map[string]int, error) {
	// /proc/meminfo has a lot of statistics (many of them aren't always useful)
	// we might as well get all of them
	data, err := ioutil.ReadFile(MemoryStatsPath)
	if err != nil {
		return map[string]int{}, err
	}
	stringList := strings.Split(string(data), "\n")

	var memStats = make(map[string]int)
	for _, line := range stringList {
		if len(line) == 0 {
			continue
		}

		tmpField, tmpValue := getMemoryStatFromLine(line)
		if tmpField == "" {
			log.Printf("Could not find field in %v\n", line)
			continue
		}
		if _, ok := memStats[tmpField]; ok {
			// this really should not happen, but just in case
			log.Printf("Field %s was already read, overwriting\n", tmpField)
		}
		memStats[tmpField] = tmpValue
	}

	return memStats, nil
}

func getMemoryStatFromLine(line string) (string, int) {
	// given a line from /proc/meminfo, return string containing the stat name and its value
	// get field name
	pos := strings.Index(line, ":")
	if pos == -1 {
		return "", -1
	}
	field := line[0:pos]

	// get value
	pos = strings.LastIndex(line[0:len(line)-3], " ")
	if pos == -1 {
		return field, -1
	}
	tmpValue := line[pos+1 : len(line)-3]
	tmpValueInt, err := strconv.ParseInt(tmpValue, 10, 32)
	if err != nil {
		log.Printf("Could not parse value in line %s\n%v", line, err)
		return field, -1
	}

	return field, int(tmpValueInt)
}

func reportMemoryStatsToInflux(dbName, piName string, stat map[string]int, now time.Time, c client.Client) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  dbName,
		Precision: "ms",
	})
	if err != nil {
		return err
	}

	tags := map[string]string{
		"pi_name": piName,
	}

	fields := map[string]interface{}{}
	for k, v := range stat {
		fields[k] = v
	}

	point, err := client.NewPoint(MemoryMeasurementsName, tags, fields, now)
	bp.AddPoint(point)
	err = c.Write(bp)
	if err != nil {
		return err
	}

	return nil
}
