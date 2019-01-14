package gcloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"golang.org/x/oauth2/google"

	"github.com/czerwonk/dns-drain/changelog"

	dns "google.golang.org/api/dns/v1"
)

type GoogleDnsUndrainer struct {
	project    string
	dryRun     bool
	zoneFilter *regexp.Regexp
	skipFilter *regexp.Regexp
	service    *dns.Service
	updater    *recordUpdater
	limit      int64
}

type groupKey struct {
	record     string
	recordType string
}

func NewUndrainer(project string, dryRun bool, zoneFilter *regexp.Regexp, skipFilter *regexp.Regexp, limit int64) *GoogleDnsUndrainer {
	return &GoogleDnsUndrainer{project: project, dryRun: dryRun, zoneFilter: zoneFilter, skipFilter: skipFilter, limit: limit}
}

func (client *GoogleDnsUndrainer) Undrain(changes *changelog.DnsChangeSet) error {
	ctx := context.Background()
	c, err := google.DefaultClient(ctx, dns.CloudPlatformScope)
	if err != nil {
		return err
	}

	client.service, err = dns.New(c)
	if err != nil {
		return err
	}

	client.updater = &recordUpdater{service: client.service, project: client.project, dryRun: client.dryRun, limit: client.limit}

	return client.undrain(changes)
}

func (client *GoogleDnsUndrainer) undrain(changes *changelog.DnsChangeSet) error {
	g := changes.GroupByZone()
	done := make(chan bool)

	for z, c := range g {
		go client.undrainZone(z, c, done)
	}

	for i := 0; i < len(g); i++ {
		select {
		case <-done:
		case <-time.After(60 * time.Second):
			return errors.New(fmt.Sprintln("Timeout exceeded!"))
		}
	}

	return nil
}

func (client *GoogleDnsUndrainer) undrainZone(zone string, changes []changelog.DnsChange, done chan bool) error {
	defer func() { done <- true }()

	if client.skipFilter != nil && client.skipFilter.MatchString(zone) {
		return nil
	}

	if client.zoneFilter != nil && !client.zoneFilter.MatchString(zone) {
		return nil
	}

	res, err := client.service.ResourceRecordSets.List(client.project, zone).Do()
	if err != nil {
		return err
	}

	for r, c := range groupChanges(changes) {
		err = client.revertChange(r.record, c, res.Rrsets)
		if err != nil {
			return err
		}
	}

	return nil
}

func groupChanges(changes []changelog.DnsChange) map[groupKey][]changelog.DnsChange {
	m := make(map[groupKey][]changelog.DnsChange)
	for _, x := range changes {
		key := groupKey{
			record:     x.Record,
			recordType: x.RecordType,
		}

		var arr []changelog.DnsChange
		var found bool
		if arr, found = m[key]; !found {
			arr = make([]changelog.DnsChange, 0)
		}

		m[key] = append(arr, x)
	}

	return m
}

func (client *GoogleDnsUndrainer) revertChange(record string, changes []changelog.DnsChange, records []*dns.ResourceRecordSet) error {
	rec := findRecordSet(record, changes[0].RecordType, records)
	if rec == nil {
		log.Printf("WARNING - Record %s not found in zone %s\n", record, changes[0].Zone)
		rec = &dns.ResourceRecordSet{
			Name:    record,
			Type:    changes[0].RecordType,
			Rrdatas: make([]string, 0),
		}
	}

	d := client.getNewDatas(changes, rec)
	return client.updateRecordSet(rec, changes[0].Zone, d)
}

func (client *GoogleDnsUndrainer) getNewDatas(changes []changelog.DnsChange, record *dns.ResourceRecordSet) []string {
	m := make(map[string]int)
	for _, x := range record.Rrdatas {
		m[x] = 1
	}

	for _, c := range changes {
		if c.Action == changelog.Add {
			delete(m, c.Value)
		} else {
			m[c.Value] = 1
		}
	}

	r := make([]string, 0)
	for k, v := range m {
		if v == 1 {
			r = append(r, k)
		}
	}

	return r
}

func findRecordSet(name, recordType string, records []*dns.ResourceRecordSet) *dns.ResourceRecordSet {
	for _, r := range records {
		if r.Name == name && r.Type == recordType {
			return r
		}
	}

	return nil
}

func (client *GoogleDnsUndrainer) updateRecordSet(rec *dns.ResourceRecordSet, zone string, datas []string) error {
	_, err := client.updater.updateRecordSet(zone, rec, datas)
	if err != nil {
		return err
	}

	return nil
}
