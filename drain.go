package main

import (
	"log"
	"net"

	"github.com/czerwonk/dns-drain/changelog"
	"github.com/czerwonk/dns-drain/gcloud"
)

func drain(ipNet *net.IPNet, newIp net.IP) error {
	logger, err := changelog.NewFileChangeLogger(*file)
	if err != nil {
		return err
	}
	defer flushAndCloseLogger(logger)

	c := gcloud.NewDrainer(*gcloudProject, *dry, zoneFilterRegex, skipFilterRegex, logger)
	return c.Drain(ipNet, newIp)
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
