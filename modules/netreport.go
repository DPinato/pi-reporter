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

const DefaultNetReportTime = 30 * time.Second
const BaseNetStatsDir = "/sys/class/net/"
const NetMeasurementsName = "network_stats"

var NetStatsList = []string{"collisions", "rx_crc_errors", "rx_frame_errors", "rx_over_errors", "tx_carrier_errors",
	"tx_fifo_errors", "multicast", "rx_dropped", "rx_length_errors", "rx_packets",
	"tx_compressed", "tx_heartbeat_errors", "rx_bytes", "rx_errors", "rx_missed_errors",
	"tx_aborted_errors", "tx_dropped", "tx_packets", "rx_compressed", "rx_fifo_errors",
	"rx_nohandler", "tx_bytes", "tx_errors", "tx_window_errors"}

type NetIFStats struct {
	IfName     string
	Speed      int64
	Statistics map[string]int64
}

func ReportNetworkStats(dbName string, c client.Client) error {
	myName, _ := helper.GetPIName(helper.PINetIfaces[0])
	log.Printf("ReportNetworkStats() is starting, %s\n", myName)

	ticker := time.NewTicker(DefaultNetReportTime)
	for {
		select {
		case t := <-ticker.C:
			for _, ifName := range helper.PINetIfaces {
				stat, err := getNetworkIfStatistics(ifName)
				if err != nil {
					log.Println(err)
				}

				err = reportNetStatsToInflux(dbName, myName, stat, t, c)
				if err != nil {
					log.Println(err)
				}

			}

		}
	}
}

func getNetworkIfStatistics(ifName string) (NetIFStats, error) {
	// get statistics for a wired interface from the Linux sysfs filesystem
	// https://man7.org/linux/man-pages/man5/sysfs.5.html
	var err error
	var sample NetIFStats
	var statsMap = make(map[string]int64)
	statsDir := BaseNetStatsDir + ifName + "/"

	stat, err := ioutil.ReadFile(statsDir + "speed")
	sample.IfName = ifName
	sample.Speed, err = strconv.ParseInt(strings.TrimSuffix(string(stat), "\n"), 10, 64)
	statsDir += "statistics/"

	// go through all the statistics
	for _, elem := range NetStatsList {
		path := statsDir + elem
		stat, err = ioutil.ReadFile(path)
		statInt, _ := strconv.ParseInt(strings.TrimSuffix(string(stat), "\n"), 10, 64)
		statsMap[elem] = statInt
	}

	sample.Statistics = statsMap
	return sample, err
}

func reportNetStatsToInflux(dbName, piName string, stat NetIFStats, now time.Time, c client.Client) error {
	tags := map[string]string{
		"pi_name": piName,
		"if_name": stat.IfName,
	}
	fields := map[string]interface{}{}
	fields["speed"] = stat.Speed
	for k, v := range stat.Statistics {
		fields[k] = v
	}

	var dbInfoObj helper.DBInfo
	dbInfoObj.DBName = dbName
	dbInfoObj.MeasName = NetMeasurementsName
	dbInfoObj.Tags = tags
	dbInfoObj.Fields = fields
	dbInfoObj.Now = now

	err := helper.ReportStatsToInflux(dbInfoObj, c)
	if err != nil {
		return err
	}
	return nil
}
