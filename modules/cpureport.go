package modules

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/dpinato/pi-reporter/helper"
	client "github.com/influxdata/influxdb1-client/v2"
)

// CPULoad contains statistics for all CPU cores, including the whole package
type CPULoad struct {
	Stats []CPUCoreLoad
}

// CPUCoreLoad contains statistics for one CPU core, and the whole package (first line of /proc/stat)
type CPUCoreLoad struct {
	User      int64
	Nice      int64
	System    int64
	Idle      int64
	IoWait    int64
	Irq       int64
	Softirq   int64
	Steal     int64
	Guest     int64
	GuestNice int64
}

const DefaultCPUReportTime = 30 * time.Second
const CPUStatsFile = "/proc/stat"
const CPUMeasurementsName = "cpu_load"

func ReportCPUUsage(dbName string, c client.Client) {
	var err error
	myName, _ := helper.GetPIName(helper.PINetIfaces[0])
	log.Printf("ReportCPUUsage() is starting, %s\n", myName)

	// get first load sample
	rawPrevStat, _ := ioutil.ReadFile(CPUStatsFile)
	prevStat := readCPUUsage(string(rawPrevStat))
	// log.Println(prevStat)

	ticker := time.NewTicker(DefaultCPUReportTime)
	for {
		select {
		case <-ticker.C:
			rawCurrStat, _ := ioutil.ReadFile(CPUStatsFile)
			currStat := readCPUUsage(string(rawCurrStat))

			// get load
			currLoad := getCPUUsage(prevStat, currStat)
			// log.Printf("%.4f\t%.4f\t%.4f\t%.4f\t%.4f\n", currLoad[0], currLoad[1], currLoad[2], currLoad[3], currLoad[4])

			// report to InfluxDB
			err = reportCPUUsageToInflux(dbName, myName, currLoad, c)
			if err != nil {
				log.Println(err)
			}

			prevStat = currStat
		}
	}
}

//
//
//
func getCPUUsage(pStat, nStat CPULoad) []float64 {
	// pStat is the previous point of CPU usage
	// pStat is the latest point of CPU usage
	// followed this response for the calculations below
	// https://stackoverflow.com/questions/23367857/accurate-calculation-of-cpu-usage-given-in-percentage-in-linux
	outputLoad := make([]float64, len(pStat.Stats))

	for i := 0; i < len(pStat.Stats); i++ {
		prevIdle := pStat.Stats[i].Idle + pStat.Stats[i].IoWait
		idle := nStat.Stats[i].Idle + nStat.Stats[i].IoWait

		prevNonIdle := pStat.Stats[i].User + pStat.Stats[i].Nice + pStat.Stats[i].System + pStat.Stats[i].Irq + pStat.Stats[i].Softirq + pStat.Stats[i].Steal
		nonIdle := nStat.Stats[i].User + nStat.Stats[i].Nice + nStat.Stats[i].System + nStat.Stats[i].Irq + nStat.Stats[i].Softirq + nStat.Stats[i].Steal

		prevTotal := prevIdle + prevNonIdle
		total := idle + nonIdle

		totald := total - prevTotal
		idled := idle - prevIdle

		outputLoad[i] = float64(totald-idled) / float64(totald)
	}

	return outputLoad
}

func readCPUUsage(data string) CPULoad {
	// read raw string from /proc/stat and return a CPULoadSample
	var loadObj CPULoad
	data = strings.Replace(data, "  ", " ", -1) // the first line may contain 2 spaces after "cpu"

	for _, line := range strings.Split(data, "\n") {
		if len(line) > 3 && line[0:3] == "cpu" {
			var tmpLoad CPUCoreLoad
			list := strings.Split(line, " ")

			tmpLoad.User, _ = strconv.ParseInt(list[1], 10, 32)
			tmpLoad.Nice, _ = strconv.ParseInt(list[2], 10, 32)
			tmpLoad.System, _ = strconv.ParseInt(list[3], 10, 32)
			tmpLoad.Idle, _ = strconv.ParseInt(list[4], 10, 32)
			tmpLoad.IoWait, _ = strconv.ParseInt(list[5], 10, 32)
			tmpLoad.Irq, _ = strconv.ParseInt(list[6], 10, 32)
			tmpLoad.Softirq, _ = strconv.ParseInt(list[7], 10, 32)
			tmpLoad.Steal, _ = strconv.ParseInt(list[8], 10, 32)
			tmpLoad.Guest, _ = strconv.ParseInt(list[9], 10, 32)
			tmpLoad.GuestNice, _ = strconv.ParseInt(list[10], 10, 32)

			loadObj.Stats = append(loadObj.Stats, tmpLoad)
		}
	}

	return loadObj
}

func reportCPUUsageToInflux(dbName, piName string, load []float64, c client.Client) error {

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
	fields["cpu"] = load[0]
	for i, elem := range load[1:] {
		key := fmt.Sprintf("cpu_%d", i)
		fields[key] = elem
	}

	point, err := client.NewPoint(CPUMeasurementsName, tags, fields, time.Now())
	bp.AddPoint(point)
	err = c.Write(bp)
	if err != nil {
		return err
	}

	return nil
}
