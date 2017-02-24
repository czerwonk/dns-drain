package gcloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"time"

	"golang.org/x/oauth2/google"

	"github.com/czerwonk/dns-drain/changelog"

	dns "google.golang.org/api/dns/v1"
)

type GoogleDnsDrainer struct {
	project    string
	dryRun     bool
	zoneFilter *regexp.Regexp
	skipFilter *regexp.Regexp
	service    *dns.Service
	logger     changelog.ChangeLogger
	updater    *recordUpdater
}

type DrainFilter func(*dns.ResourceRecordSet) []string

func NewDrainer(project string, dryRun bool, zoneFilter *regexp.Regexp, skipFilter *regexp.Regexp, changelogger changelog.ChangeLogger) *GoogleDnsDrainer {
	return &GoogleDnsDrainer{project: project, dryRun: dryRun, zoneFilter: zoneFilter, skipFilter: skipFilter, logger: changelogger}
}

func (client *GoogleDnsDrainer) DrainWithIpNet(ipNet *net.IPNet, newIp net.IP) error {
	filter := func(rec *dns.ResourceRecordSet) []string {
		return filterWithIpNet(rec, ipNet)
	}

	newValue := ""
	if newIp != nil {
		newValue = newIp.String()
	}

	return client.performForZones(filter, newValue)
}

func (client *GoogleDnsDrainer) DrainWithValue(value string, newValue string) error {
	filter := func(rec *dns.ResourceRecordSet) []string {
		return filterWithValue(rec, value)
	}

	return client.performForZones(filter, newValue)
}

func (client *GoogleDnsDrainer) performForZones(filter DrainFilter, newValue string) error {
	ctx := context.Background()
	c, err := google.DefaultClient(ctx, dns.CloudPlatformScope)
	if err != nil {
		return err
	}

	client.service, err = dns.New(c)
	if err != nil {
		return err
	}

	client.updater = &recordUpdater{service: client.service, project: client.project, dryRun: client.dryRun}

	zones, err := client.getZones()
	if err != nil {
		return err
	}

	done := make(chan bool)
	for _, z := range zones {
		go client.drainForZone(z.Name, filter, newValue, done)
	}

	for i := 0; i < len(zones); i++ {
		select {
		case <-done:
		case <-time.After(60 * time.Second):
			return errors.New(fmt.Sprintln("Timeout exceeded!"))
		}
	}

	return nil
}

func (client *GoogleDnsDrainer) getZones() ([]*dns.ManagedZone, error) {
	r, err := client.service.ManagedZones.List(client.project).Do()
	if err != nil {
		return nil, err
	}

	zones := make([]*dns.ManagedZone, 0)
	for _, z := range r.ManagedZones {
		if !client.matchesSkipFilter(z.Name) && client.matchesZoneFilter(z.Name) {
			zones = append(zones, z)
		}
	}

	return zones, nil
}

func (client *GoogleDnsDrainer) matchesSkipFilter(zone string) bool {
	return client.skipFilter != nil && client.skipFilter.MatchString(zone)
}

func (client *GoogleDnsDrainer) matchesZoneFilter(zone string) bool {
	return client.zoneFilter == nil || client.zoneFilter.MatchString(zone)
}

func (client *GoogleDnsDrainer) drainForZone(zone string, filter DrainFilter, newValue string, done chan bool) {
	defer func() { done <- true }()

	r, err := client.service.ResourceRecordSets.List(client.project, zone).Do()
	if err != nil {
		log.Printf("ERROR - %s: %s\n", zone, err)
		return
	}

	for _, rec := range r.Rrsets {
		client.handleRecordSet(zone, rec, newValue, filter)
	}
}

func (client *GoogleDnsDrainer) handleRecordSet(zone string, rec *dns.ResourceRecordSet, newValue string, filter DrainFilter) {
	d := filter(rec)

	if len(d) == 0 && len(newValue) == 0 {
		log.Printf("WARN - %s %s: Only one value assigned to record. Can not drain!\n", rec.Type, rec.Name)
		return
	}

	if len(d) == len(rec.Rrdatas) {
		return
	}

	if len(newValue) > 0 && !isInDatas(newValue, d) {
		d = append(d, newValue)
	}

	err := client.updateRecordSet(rec, zone, d)
	if err != nil {
		log.Printf("ERROR - %s: %s", rec.Name, err)
	}
}

func filterWithValue(rec *dns.ResourceRecordSet, value string) []string {
	res := make([]string, 0)

	for _, x := range rec.Rrdatas {
		if x != value {
			res = append(res, x)
		}
	}

	return res
}

func filterWithIpNet(rec *dns.ResourceRecordSet, ipNet *net.IPNet) []string {
	res := make([]string, 0)

	for _, x := range rec.Rrdatas {
		ip := net.ParseIP(x)
		if ip == nil || !ipNet.Contains(ip) {
			res = append(res, x)
		}
	}

	return res
}

func isInDatas(value string, datas []string) bool {
	for _, x := range datas {
		if x == value {
			return true
		}
	}

	return false
}

func (client *GoogleDnsDrainer) updateRecordSet(rec *dns.ResourceRecordSet, zone string, datas []string) error {
	err := client.updater.updateRecordSet(zone, rec, datas)
	if err != nil {
		return err
	}

	return client.logChanges(rec, zone, datas)
}

func (client *GoogleDnsDrainer) logChanges(rec *dns.ResourceRecordSet, zone string, datas []string) error {
	before := rec.Rrdatas
	after := datas

	m := make(map[string]int)
	for _, x := range before {
		m[x] = -1
	}
	for _, y := range after {
		m[y] += 1
	}

	for k, v := range m {
		if v != 0 {
			err := client.logChange(rec, zone, k, v)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (client *GoogleDnsDrainer) logChange(rec *dns.ResourceRecordSet, zone string, value string, changeValue int) error {
	var action string
	if changeValue == 1 {
		action = changelog.Add
	} else {
		action = changelog.Remove
	}

	c := changelog.DnsChange{Provider: "gcloud", Zone: zone, Record: rec.Name, RecordType: rec.Type, Value: value, Action: action}
	return client.logger.LogChange(c)
}
