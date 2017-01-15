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
}

func NewDrainer(project string, dryRun bool, zoneFilter *regexp.Regexp, skipFilter *regexp.Regexp, changelogger changelog.ChangeLogger) *GoogleDnsDrainer {
	return &GoogleDnsDrainer{project: project, dryRun: dryRun, zoneFilter: zoneFilter, skipFilter: skipFilter, logger: changelogger}
}

func (client *GoogleDnsDrainer) Drain(ipNet *net.IPNet, newIp net.IP) error {
	ctx := context.Background()
	c, err := google.DefaultClient(ctx, dns.CloudPlatformScope)
	if err != nil {
		return err
	}

	client.service, err = dns.New(c)
	if err != nil {
		return err
	}

	zones, err := client.getZones()
	if err != nil {
		return err
	}

	done := make(chan bool)
	for _, z := range zones {
		go client.drainWithIpNet(z.Name, ipNet, newIp, done)
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

func (client *GoogleDnsDrainer) drainWithIpNet(zone string, ipNet *net.IPNet, newIp net.IP, done chan bool) {
	defer func() { done <- true }()

	r, err := client.service.ResourceRecordSets.List(client.project, zone).Do()
	if err != nil {
		log.Printf("ERROR - %s: %s\n", zone, err)
		return
	}

	for _, rec := range r.Rrsets {
		client.handleRecordSet(zone, rec, ipNet, newIp)
	}
}

func (client *GoogleDnsDrainer) handleRecordSet(zone string, rec *dns.ResourceRecordSet, ipNet *net.IPNet, newIp net.IP) {
	d := filterWithIpNet(rec, ipNet)

	if len(d) == 0 && newIp == nil {
		log.Printf("WARN - %s %s: Only one IP assigned to record. Can not drain!\n", rec.Type, rec.Name)
		return
	}

	if len(d) == len(rec.Rrdatas) {
		return
	}

	if newIp != nil && !isInDatas(newIp, d) {
		d = append(d, newIp.String())
	}

	err := client.updateRecordSet(rec, zone, d)
	if err != nil {
		log.Printf("ERROR - %s: %s", rec.Name, err)
	}
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

func isInDatas(ip net.IP, datas []string) bool {
	for _, x := range datas {
		if x == ip.String() {
			return true
		}
	}

	return false
}

func (client *GoogleDnsDrainer) updateRecordSet(rec *dns.ResourceRecordSet, zone string, datas []string) error {
	c := &updateContext{service: client.service, project: client.project, zone: zone, dryRun: client.dryRun}
	err := updateRecordSet(rec, datas, c)
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
