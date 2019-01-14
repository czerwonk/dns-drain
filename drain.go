package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"

	"github.com/czerwonk/dns-drain/changelog"
	"github.com/czerwonk/dns-drain/gcloud"
)

type DrainActionFunc func(Drainer) error

func drain() error {
	if modeIsAmbigous() {
		return errors.New("Mode is ambigous. Please use either ip, value or regex parameter.")
	}

	if len(*value) > 0 {
		return drainWithValue()
	}

	if len(*regexString) > 0 {
		return drainWithRegex()
	}

	return drainWithIpNet()
}

func modeIsAmbigous() bool {
	i := 0

	if len(*value) > 0 {
		i++
	}

	if len(*regexString) > 0 {
		i++
	}

	if len(*ip) > 0 {
		i++
	}

	return i != 1
}

func drainWithRegex() error {
	regex, err := regexp.Compile(*regexString)
	if err != nil {
		return err
	}

	actionFunc := func(d Drainer) error {
		return d.DrainWithRegex(regex, *newValue)
	}
	return performDrain(actionFunc)
}

func drainWithValue() error {
	actionFunc := func(d Drainer) error {
		return d.DrainWithValue(*value, *newValue)
	}
	return performDrain(actionFunc)
}

func drainWithIpNet() error {
	ipNet, err := getNetFromIp()
	if err != nil {
		return err
	}

	var newIp net.IP
	if len(*newIpStr) > 0 {
		newIp = net.ParseIP(*newIpStr)
	}

	actionFunc := func(d Drainer) error {
		return d.DrainWithIpNet(ipNet, newIp)
	}
	return performDrain(actionFunc)
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

func performDrain(actionFunc DrainActionFunc) error {
	logger, err := changelog.NewFileChangeLogger(*file)
	if err != nil {
		return err
	}
	defer flushAndCloseLogger(logger)

	c := gcloud.NewDrainer(*gcloudProject, *dry, zoneFilterRegex, skipFilterRegex, *typeFilter, logger, *force, *limit)
	return actionFunc(c)
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
