package main

import (
	"log"
	"net"
	"time"

	"github.com/czerwonk/dns-drain/changelog"
	"github.com/czerwonk/dns-drain/gcloud"
)

func drain(ipNet *net.IPNet, newIp net.IP) error {
	logger, err := changelog.NewFileChangeLogger(*file)
	if err != nil {
		return err
	}
	defer flushAndCloseLogger(logger)

	start := time.Now()

	c := gcloud.NewDrainer(*gcloudProject, *dry, zoneFilterRegex, logger)
	err = c.Drain(ipNet, newIp)

	if err == nil {
		log.Printf("Finished after %v", time.Since(start))
	}

	return err
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
