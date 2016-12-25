package gcloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/oauth2/google"

	"github.com/czerwonk/dns-drain/changelog"

	dns "google.golang.org/api/dns/v1"
)

type GoogleDnsClient struct {
	Project    string
	DryRun     bool
	ZoneFilter string
	service    *dns.Service
	logger     changelog.ChangeLogger
}

func NewClient(project string, dryRun bool, zoneFilter string, changelogger changelog.ChangeLogger) *GoogleDnsClient {
	return &GoogleDnsClient{Project: project, DryRun: dryRun, ZoneFilter: zoneFilter, logger: changelogger}
}

func (client *GoogleDnsClient) Drain(ipNet *net.IPNet, newIp net.IP) error {
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

func (client *GoogleDnsClient) getZones() ([]*dns.ManagedZone, error) {
	r, err := client.service.ManagedZones.List(client.Project).Do()
	if err != nil {
		return nil, err
	}

	zones := make([]*dns.ManagedZone, 0)
	for _, z := range r.ManagedZones {
		if len(client.ZoneFilter) == 0 || z.Name == client.ZoneFilter {
			zones = append(zones, z)
		}
	}

	return zones, nil
}

func (client *GoogleDnsClient) drainWithIpNet(zone string, ipNet *net.IPNet, newIp net.IP, done chan bool) {
	defer func() { done <- true }()

	r, err := client.service.ResourceRecordSets.List(client.Project, zone).Do()
	if err != nil {
		log.Printf("ERROR - %s: %s\n", zone, err)
		return
	}

	for _, rec := range r.Rrsets {
		client.handleRecordSet(zone, rec, ipNet, newIp)
	}
}

func (client *GoogleDnsClient) handleRecordSet(zone string, rec *dns.ResourceRecordSet, ipNet *net.IPNet, newIp net.IP) {
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

func (client *GoogleDnsClient) updateRecordSet(rec *dns.ResourceRecordSet, zone string, datas []string) error {
	log.Printf("- %s: %s %s\n", rec.Name, rec.Type, rec.Rrdatas)
	log.Printf("+ %s: %s %s\n", rec.Name, rec.Type, datas)

	if client.DryRun {
		return nil
	}

	c := &dns.Change{Additions: make([]*dns.ResourceRecordSet, 0), Deletions: make([]*dns.ResourceRecordSet, 0)}
	c.Deletions = append(c.Deletions, rec)

	updated := *rec
	updated.Rrdatas = datas
	c.Additions = append(c.Additions, &updated)

	_, err := client.service.Changes.Create(client.Project, zone, c).Do()
	if err != nil {
		return err
	}

	return client.logChanges(rec, zone, datas)
}

func (client *GoogleDnsClient) logChanges(rec *dns.ResourceRecordSet, zone string, datas []string) error {
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

func (client *GoogleDnsClient) logChange(rec *dns.ResourceRecordSet, zone string, value string, changeValue int) error {
	var action string
	if changeValue == 1 {
		action = changelog.Add
	} else {
		action = changelog.Remove
	}

	c := changelog.DnsChange{Zone: zone, Record: rec.Name, RecordType: rec.Type, Value: value, Action: action}
	return client.logger.LogChange("gcloud", c)
}
