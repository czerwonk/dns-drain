package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"time"
)

const version string = "0.5.1"

var (
	showVersion     = flag.Bool("version", false, "Show version information")
	ip              = flag.String("ip", "", "IP or net address to remove from DNS")
	newIpStr        = flag.String("new_ip", "", "IP to set instead of removed IP")
	gcloudProject   = flag.String("gcloud.project", "", "Project ID for Google Cloud DNS")
	dry             = flag.Bool("dry", false, "Do not modify DNS records (simulation only)")
	zoneFilter      = flag.String("zone", "", "Apply only to zones matching the specifed regex")
	skipFilter      = flag.String("skip", "", "Skip zones matching the specified regex")
	file            = flag.String("file", "drain.json", "File containing changes (for log or undrain)")
	shouldUndrain   = flag.Bool("undrain", false, "Use file to revert changes")
	zoneFilterRegex *regexp.Regexp
	skipFilterRegex *regexp.Regexp
)

func main() {
	flag.Parse()

	if *showVersion {
		printVersionInfo()
		return
	}

	err := parseFilterArgs()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if *dry {
		log.Println("Using dry run. No records will be changed.")
	}

	start := time.Now()

	if *shouldUndrain {
		err = undrain(*file)
	} else {
		err = startDrain()
	}

	if err != nil {
		log.Printf("ERROR - %s\n", err)
		os.Exit(1)
	}

	log.Printf("Finished after %v", time.Since(start))
}

func printVersionInfo() {
	fmt.Println("dns-drain")
	fmt.Printf("Version: %s\n", version)
}

func parseFilterArgs() error {
	var err error = nil

	if len(*zoneFilter) > 0 {
		zoneFilterRegex, err = regexp.Compile(*zoneFilter)
		if err != nil {
			return err
		}
	}

	if len(*skipFilter) > 0 {
		skipFilterRegex, err = regexp.Compile(*skipFilter)
		if err != nil {
			return err
		}
	}

	return nil
}

func startDrain() error {
	_, ipNet, err := net.ParseCIDR(*ip)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if ipNet == nil {
		return errors.New("Please use CIDR notation")
	}

	var newIp net.IP
	if len(*newIpStr) > 0 {
		newIp = net.ParseIP(*newIpStr)
	}

	return drain(ipNet, newIp)
}
