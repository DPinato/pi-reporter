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

// SupportedArgs:
// --env: Indicates the environment type, i.e. dev or prod
// --influxhost: IP address of the host running InfluxDB
var SupportedArgs = []string{"--env", "--influxhost"}

// constants for InfluxDB connection
const (
	InfluxDBPort     = "8086"
	InfluxDBNameProd = "pi_reporter_prod"
	InfluxDBNameDev  = "pi_reporter_dev"
)

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

	// check input arguments
	args := parseCmdArgs(os.Args[1:])
	ok := validateCmdArgs(args)
	if !ok || len(args) < 2 {
		log.Fatalf("Bad or missing command arguments, %v\n", args)
	}
	log.Println(args)

	// initialise things for the environment selected
	var influxDBName string
	switch args["--env"] {
	case "dev":
		influxDBName = InfluxDBNameDev
	case "prod":
		influxDBName = InfluxDBNameProd
	default:
		log.Fatalln("Bad environment selected, terminating ...")
	}

	// connect to InfluxDB
	influxDBHost := args["--influxhost"]
	c, err := influxDBClient(influxDBHost, InfluxDBPort)
	if err != nil {
		log.Println("Error creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()
	log.Printf("Connected to DB %s:%s\n", influxDBHost, InfluxDBPort)

	// start reporting
	var wg sync.WaitGroup

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		// these routines should all run at the same time anyway
		// TODO: should use a channel to listen for an error that breaks the routine
		defer wg.Done()
		go modules.ReportCPUUsage(influxDBName, c)
		go modules.ReportNetworkStats(influxDBName, c)
		go modules.ReportTempStats(influxDBName, c)
		go modules.ReportMemoryStats(influxDBName, c)
		modules.ReportDiskStats(influxDBName, c)
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

func parseCmdArgs(args []string) map[string]string {
	// a key will start with --, the value should be right after the key
	output := make(map[string]string)
	if len(args) != 0 && len(args)%2 == 0 {
		for i := 0; i < len(args); i += 2 {
			output[args[i]] = args[i+1]
		}
	}

	return output
}

func validateCmdArgs(args map[string]string) bool {
	// return true if the arguments provided are within the list of supported arguments
	for k, v := range args {
		var found bool
		for _, elem := range SupportedArgs {
			if k == elem && v != "" {
				found = true
				continue
			}
		}

		if !found {
			// not in list of supported arguments
			return false
		}
	}

	return true
}
