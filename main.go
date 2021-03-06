package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"
)

const version string = "0.8.2"

var (
	showVersion     = flag.Bool("version", false, "Show version information")
	ip              = flag.String("ip", "", "IP or net address to remove from DNS")
	newIpStr        = flag.String("new_ip", "", "IP to set instead of removed IP")
	gcloudProject   = flag.String("gcloud.project", "", "Project ID for Google Cloud DNS")
	dry             = flag.Bool("dry", false, "Do not modify DNS records (simulation only)")
	zoneFilter      = flag.String("zone", "", "Apply only to zones matching the specified regex")
	skipFilter      = flag.String("skip", "", "Skip zones matching the specified regex")
	typeFilter      = flag.String("type", "", "Record type to change")
	file            = flag.String("file", "drain.json", "File containing changes (for log or undrain)")
	shouldUndrain   = flag.Bool("undrain", false, "Use file to revert changes")
	value           = flag.String("value", "", "Value to replace/replace in DNS data")
	regexString     = flag.String("regex", "", "Regex to find data in DNS records to remove/replace")
	newValue        = flag.String("new_value", "", "Value to replace with in DNS data")
	force           = flag.Bool("force", false, "Remove value from record even if it is the only value")
	limit           = flag.Int64("limit", -1, "Max number of records to change (-1 = unlimited)")
	zoneFilterRegex *regexp.Regexp
	skipFilterRegex *regexp.Regexp
)

func init() {
	flag.Usage = func() { printUsage() }
}

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

	if *force {
		log.Println("Logic check was disabled. There is no guarantee for a consistent result.")
	}

	start := time.Now()

	if *shouldUndrain {
		err = undrain(*file)
	} else {
		err = drain()
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
	fmt.Println("Author(s): Daniel Czerwonk")
	fmt.Println("Drain and undrain frontends by using DNS")
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
