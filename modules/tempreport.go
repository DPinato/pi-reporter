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
	tags := map[string]string{
		"pi_name": piName,
	}
	fields := map[string]interface{}{}
	fields["temperature"] = stat

	var dbInfoObj helper.DBInfo
	dbInfoObj.DBName = dbName
	dbInfoObj.MeasName = TempMeasurementsName
	dbInfoObj.Tags = tags
	dbInfoObj.Fields = fields
	dbInfoObj.Now = now

	err := helper.ReportStatsToInflux(dbInfoObj, c)
	if err != nil {
		return err
	}
	return nil
}
