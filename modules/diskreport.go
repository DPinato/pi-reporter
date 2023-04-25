package modules

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dpinato/pi-reporter/helper"
	client "github.com/influxdata/influxdb1-client/v2"
)

const DefaultDiskReportTime = 30 * time.Second
const DiskStatsPath = "/proc/diskstats"
const DiskMeasurementsName = "disk_stats"
const DiskNameRegexp = "sd|mmcblk"

// DiskStats is based on reading the fields in /proc/diskstats
// /proc/diskstats format: https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats
// more: https://www.kernel.org/doc/Documentation/block/stat.txt
type DiskStats struct {
	DevName string `json:"device_name"`

	ReadIOs        int64 `json:"read_ios"` // reads completed successfully
	ReadMerges     int64 `json:"read_merges"`
	ReadSectors    int64 `json:"read_sectors"`
	ReadTicks      int64 `json:"read_ticks"` // time spent reading (ms)
	WriteIOs       int64 `json:"write_ios"`  // writes completed
	WriteMerges    int64 `json:"write_merges"`
	WriteSectors   int64 `json:"write_sectors"`
	WriteTicks     int64 `json:"write_ticks"`   // time spent writing (ms)
	InFlight       int64 `json:"in_flight"`     // I/Os currently in progress
	IoTicks        int64 `json:"io_ticks"`      // time spent doing I/Os (ms)
	TimeInQueue    int64 `json:"time_in_queue"` // weighted time spent doing I/Os (ms)
	DiscardIOs     int64 `json:"discard_ios"`
	DiscardMerges  int64 `json:"discard_merges"`
	DiscardSectors int64 `json:"discard_sectors"`
	DiscardTicks   int64 `json:"discard_ticks"`       // time spent discarding (ms)
	FlushSuccess   int64 `json:"flush_success_count"` // flush requests completed successfully
	FlushingTicks  int64 `json:"flushing_ticks"`      // time spent flushing (ms)
}

func ReportDiskStats(dbName string, c client.Client) error {
	myName, _ := helper.GetPIName(helper.PINetIfaces[0])
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Failed to get hostname - %+v", err)
		hostname = myName
	} else if hostname == helper.PIDefaultHostname {
		hostname = myName
	}

	log.Printf("ReportDiskStats() is starting, %s\n", hostname)

	r, err := regexp.Compile(DiskNameRegexp)
	if err != nil {
		log.Println("DiskNameRegexp is invalid")
		return err

	}

	ticker := time.NewTicker(DefaultDiskReportTime)
	for {
		select {
		case t := <-ticker.C:
			stats, err := getDiskStats(r)
			if err != nil {
				log.Println(err)
				continue
			}

			for _, elem := range stats {
				err = reportDiskStatsToInflux(dbName, hostname, elem, t, c)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func getDiskStats(driveReg *regexp.Regexp) (map[string]DiskStats, error) {
	// read /proc/diskstats and return disk statistics
	output := map[string]DiskStats{}

	r, err := regexp.Compile(DiskNameRegexp)
	if err != nil {
		return nil, fmt.Errorf("error while parsing regexp, %v", err)
	}

	statsBytes, err := ioutil.ReadFile(DiskStatsPath)
	if err != nil {
		return nil, err
	}

	// read this line by line so we can ignore partitions that are not relevant
	scanner := bufio.NewScanner(strings.NewReader(string(statsBytes)))
	for scanner.Scan() {
		line := scanner.Text()
		regMatchIndex := r.FindStringSubmatchIndex(line)
		if len(regMatchIndex) > 0 {
			formattedLine := formatDiskStatsLine(line)
			stats := getDiskStatsFromLine(formattedLine)
			output[stats.DevName] = stats
		}
	}

	return output, nil
}

func formatDiskStatsLine(sampleLine string) string {
	// remove all spaces at the start of the line and any 2 or more consecutive spaces
	regStr1 := "^\\s{1,}"
	regStr2 := "\\s{2,}"
	output := sampleLine

	r1, _ := regexp.Compile(regStr1)
	r2, _ := regexp.Compile(regStr2)

	output = r1.ReplaceAllString(output, "")
	output = r2.ReplaceAllString(output, " ")

	return output
}

func getDiskStatsFromLine(sampleLine string) DiskStats {
	// given a line of /proc/diskstats, return an object containing the statistics
	outStats := DiskStats{}

	list := strings.Split(sampleLine, " ")

	outStats.DevName = list[2]
	outStats.ReadIOs, _ = strconv.ParseInt(list[3], 10, 64)
	outStats.ReadMerges, _ = strconv.ParseInt(list[4], 10, 64)
	outStats.ReadSectors, _ = strconv.ParseInt(list[5], 10, 64)
	outStats.ReadTicks, _ = strconv.ParseInt(list[6], 10, 64)
	outStats.WriteIOs, _ = strconv.ParseInt(list[7], 10, 64)
	outStats.WriteMerges, _ = strconv.ParseInt(list[8], 10, 64)
	outStats.WriteSectors, _ = strconv.ParseInt(list[9], 10, 64)
	outStats.WriteTicks, _ = strconv.ParseInt(list[10], 10, 64)
	outStats.InFlight, _ = strconv.ParseInt(list[11], 10, 64)
	outStats.IoTicks, _ = strconv.ParseInt(list[12], 10, 64)
	outStats.TimeInQueue, _ = strconv.ParseInt(list[13], 10, 64)
	outStats.DiscardIOs, _ = strconv.ParseInt(list[14], 10, 64)
	outStats.DiscardMerges, _ = strconv.ParseInt(list[15], 10, 64)
	outStats.DiscardSectors, _ = strconv.ParseInt(list[16], 10, 64)
	outStats.DiscardTicks, _ = strconv.ParseInt(list[17], 10, 64)
	outStats.FlushSuccess, _ = strconv.ParseInt(list[18], 10, 64)
	outStats.FlushingTicks, _ = strconv.ParseInt(list[19], 10, 64)

	return outStats
}

func reportDiskStatsToInflux(dbName, piName string, stat DiskStats, now time.Time, c client.Client) error {
	tags := map[string]string{
		"pi_name":     piName,
		"device_name": stat.DevName,
	}

	fields := map[string]interface{}{}
	v := reflect.ValueOf(stat)
	typeOfS := v.Type()
	for i := 1; i < v.NumField(); i++ {
		// skip DevName
		fieldName := typeOfS.Field(i).Name
		fieldVal := v.Field(i).Interface()
		fields[fieldName] = fieldVal
	}

	var dbInfoObj helper.DBInfo
	dbInfoObj.DBName = dbName
	dbInfoObj.MeasName = DiskMeasurementsName
	dbInfoObj.Tags = tags
	dbInfoObj.Fields = fields
	dbInfoObj.Now = now

	err := helper.ReportStatsToInflux(dbInfoObj, c)
	if err != nil {
		return err
	}
	return nil
}
