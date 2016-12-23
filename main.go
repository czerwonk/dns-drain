package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/czerwonk/dns-drain/gcloud"
)

const version string = "0.1"

var (
	showVersion   = flag.Bool("version", false, "Show version information")
	ip            = flag.String("ip", "", "IP or net address to remove from DNS")
	gcloudProject = flag.String("gcloud.project", "", "Project ID for Google Cloud DNS")
	dry           = flag.Bool("dry", false, "Do not modify DNS records (simulation only)")
)

func main() {
	flag.Parse()

	if *showVersion {
		printVersionInfo()
		os.Exit(0)
	}

	_, ipNet, err := net.ParseCIDR(*ip)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if ipNet == nil {
		log.Println("Please use CIDR notation")
	}

	err = drain(ipNet)
	if err != nil {
		log.Printf("ERROR - %s\n", err)
		os.Exit(1)
	}
}

func printVersionInfo() {
	fmt.Println("dns-drain")
	fmt.Printf("Version: %s\n", version)
}

func drain(ipNet *net.IPNet) error {
	if *dry {
		log.Println("Using dry run. No record will be changed.")
	}

	start := time.Now()

	c := gcloud.NewClient(*gcloudProject, *dry)
	err := c.Drain(ipNet)

	if err == nil {
		log.Printf("Finished after %v", time.Since(start))
	}

	return err
}
