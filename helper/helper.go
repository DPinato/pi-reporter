package helper

import (
	"net"
	"strings"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
)

var PINetIfaces = []string{"eth0", "wlan0"}
var PIDefaultHostname = "raspberrypi"

type DBInfo struct {
	DBName   string
	MeasName string
	Tags     map[string]string
	Fields   map[string]interface{}
	Now      time.Time
}

// GetPIName returns the name of the current PI using the interface name specified
// name will be something like pi-<mac>
func GetPIName(ifName string) (string, error) {
	outputStr := "pi-"
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		return "", err
	}

	macStr := iface.HardwareAddr.String()
	macStr = strings.ReplaceAll(macStr, ":", "")
	outputStr += macStr
	return outputStr, err
}

// ReportStatsToInflux reports generic statistics to InfluxDB instance using the information
// provided through the DBInfo struct
func ReportStatsToInflux(dbInfo DBInfo, c client.Client) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  dbInfo.DBName,
		Precision: "ms",
	})
	if err != nil {
		return err
	}

	point, err := client.NewPoint(dbInfo.MeasName, dbInfo.Tags, dbInfo.Fields, dbInfo.Now)
	bp.AddPoint(point)
	err = c.Write(bp)
	if err != nil {
		return err
	}
	return nil
}
