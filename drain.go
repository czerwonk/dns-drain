package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/czerwonk/dns-drain/changelog"
	"github.com/czerwonk/dns-drain/gcloud"
)

func drain() error {
	ipNet, err := getNetFromIp()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	var newIp net.IP
	if len(*newIpStr) > 0 {
		newIp = net.ParseIP(*newIpStr)
	}

	return drainByIpNet(ipNet, newIp)
}

func getNetFromIp() (*net.IPNet, error) {
	_, ipNet, err := net.ParseCIDR(*ip)

	if err != nil {
		ipAddr := net.ParseIP(*ip)
		if len(ipAddr) == 0 {
			return nil, err
		}

		var e error
		if ipAddr.To4() != nil {
			_, ipNet, e = net.ParseCIDR(fmt.Sprintf("%s/32", ipAddr))
		} else {
			_, ipNet, e = net.ParseCIDR(fmt.Sprintf("%s/128", ipAddr))
		}

		if e == nil {
			err = nil
		}
	}

	return ipNet, err
}

func drainByIpNet(ipNet *net.IPNet, newIp net.IP) error {
	logger, err := changelog.NewFileChangeLogger(*file)
	if err != nil {
		return err
	}
	defer flushAndCloseLogger(logger)

	c := gcloud.NewDrainer(*gcloudProject, *dry, zoneFilterRegex, skipFilterRegex, logger)
	return c.DrainWithIpNet(ipNet, newIp)
}

func flushAndCloseLogger(logger *changelog.FileChangeLogger) {
	err := logger.Flush()
	if err != nil {
		log.Printf("ERROR - %s\n", err)
	}

	err = logger.Close()
	if err != nil {
		log.Printf("ERROR - %s\n", err)
	}
}
