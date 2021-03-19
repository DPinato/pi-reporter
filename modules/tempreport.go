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

const DefaultTempReportTime = 30 * time.Second
const TempStatsPath = "/sys/class/thermal/thermal_zone0/temp"
const TempMeasurementsName = "temperature_stats"

func ReportTempStats(dbName string, c client.Client) error {
	myName, _ := helper.GetPIName(helper.PINetIfaces[0])
	log.Printf("ReportTempStats() is starting, %s\n", myName)

	ticker := time.NewTicker(DefaultTempReportTime)
	for {
		select {
		case t := <-ticker.C:
			stat, err := getPITemperature()
			if err != nil {
				log.Println(err)
				continue
			}

			err = reportTempStatsToInflux(dbName, myName, stat, t, c)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func getPITemperature() (float64, error) {
	stat, err := ioutil.ReadFile(TempStatsPath)
	if err != nil {
		return 0, err
	}

	tmpFloat, err := strconv.ParseFloat(strings.TrimSuffix(string(stat), "\n"), 64)
	if err != nil {
		return 0, err
	}

	return (tmpFloat / 1000.0), err
}

func reportTempStatsToInflux(dbName, piName string, stat float64, now time.Time, c client.Client) error {
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
	fields["temperature"] = stat

	point, err := client.NewPoint(TempMeasurementsName, tags, fields, now)
	bp.AddPoint(point)
	err = c.Write(bp)
	if err != nil {
		return err
	}

	return nil
}
