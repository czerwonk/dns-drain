package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

const version string = "0.3"

var (
	showVersion   = flag.Bool("version", false, "Show version information")
	ip            = flag.String("ip", "", "IP or net address to remove from DNS")
	newIpStr      = flag.String("new_ip", "", "IP to set instead of removed IP")
	gcloudProject = flag.String("gcloud.project", "", "Project ID for Google Cloud DNS")
	dry           = flag.Bool("dry", false, "Do not modify DNS records (simulation only)")
	zoneFilter    = flag.String("zone", "", "Apply only on specific zone")
	file          = flag.String("file", "drain.json", "File containing changes (for log or undrain)")
	shouldUndrain = flag.Bool("undrain", false, "Use file to revert changes")
)

func main() {
	flag.Parse()

	if *showVersion {
		printVersionInfo()
		return
	}

	if *shouldUndrain {
		err := undrain(*file)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		return
	}

	_, ipNet, err := net.ParseCIDR(*ip)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if ipNet == nil {
		log.Println("Please use CIDR notation")
	}

	var newIp net.IP
	if len(*newIpStr) > 0 {
		newIp = net.ParseIP(*newIpStr)
	}

	err = drain(ipNet, newIp)
	if err != nil {
		log.Printf("ERROR - %s\n", err)
		os.Exit(1)
	}
}

func printVersionInfo() {
	fmt.Println("dns-drain")
	fmt.Printf("Version: %s\n", version)
}
