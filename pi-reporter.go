package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/dpinato/pi-reporter/modules"
	client "github.com/influxdata/influxdb1-client/v2"
)

const LogFilePath = "/var/log/pi-reporter.log"

const InfluxDBHost = "192.168.128.200"
const InfluxDBPort = "8086"
const InfluxDBNameProd = "pi_reporter"
const InfluxDBNameDev = "pi_reporter_dev"

func main() {
	// open log file to append
	f, err := os.OpenFile(LogFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	log.Printf("pi-reporter is starting ...\n")

	// connect to InfluxDB
	c, err := influxDBClient(InfluxDBHost, InfluxDBPort)
	if err != nil {
		log.Println("Error creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()
	log.Printf("Connected to DB %s:%s\n", InfluxDBHost, InfluxDBPort)

	// start reporting
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		modules.ReportCPUUsage(InfluxDBNameDev, c)
	}(&wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		modules.ReportNetworkStats(InfluxDBNameDev, c)
	}(&wg)

	wg.Wait()
	log.Printf("pi-reporter is ending ...\n")

}

func influxDBClient(host, port string) (client.Client, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://" + host + ":" + port,
	})
	return c, err
}
