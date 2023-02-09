// SPDX-FileCopyrightText: (c) 2016 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package gcloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/czerwonk/dns-drain/pkg/changelog"
	"github.com/czerwonk/dns-drain/pkg/undrain"
	"github.com/pkg/errors"

	dns "google.golang.org/api/dns/v1"
)

type GoogleDnsUndrainer struct {
	project string
	opt     *undrain.Options
	service *dns.Service
	updater *recordUpdater
}

type groupKey struct {
	record     string
	recordType string
}

func NewUndrainer(project string, opt *undrain.Options) *GoogleDnsUndrainer {
	return &GoogleDnsUndrainer{
		project: project,
		opt:     opt,
	}
}

func (client *GoogleDnsUndrainer) Undrain(changes *changelog.DnsChangeSet) error {
	ctx := context.Background()
	svc, err := dns.NewService(ctx)
	if err != nil {
		return err
	}
	client.service = svc

	client.updater = &recordUpdater{
		service: client.service,
		project: client.project,
		dryRun:  client.opt.DryRun,
		limit:   client.opt.Limit,
	}

	return client.undrain(changes)
}

func (client *GoogleDnsUndrainer) undrain(changes *changelog.DnsChangeSet) error {
	g := changes.GroupByZone()
	doneCh := make(chan bool)
	defer close(doneCh)

	for z, c := range g {
		go client.undrainZone(z, c, doneCh)
	}

	var err error
	for i := 0; i < len(g); i++ {
		select {
		case <-doneCh:
		case <-time.After(2 * time.Minute):
			return errors.New(fmt.Sprintln("Timeout exceeded!"))
		}
	}

	return err
}

func (client *GoogleDnsUndrainer) undrainZone(zone string, changes []changelog.DnsChange, doneCh chan bool) {
	defer func() { doneCh <- true }()

	if client.opt.SkipFilter != nil && client.opt.SkipFilter.MatchString(zone) {
		return
	}

	if client.opt.ZoneFilter != nil && !client.opt.ZoneFilter.MatchString(zone) {
		return
	}

	res, err := client.service.ResourceRecordSets.List(client.project, zone).Do()
	if err != nil {
		log.Printf("ERROR - %s: %s\n", zone, err)
		return
	}

	for r, c := range groupChanges(changes) {
		err = client.revertChange(r.record, c, res.Rrsets)
		if err != nil {
			log.Printf("ERROR - %s: %s\n", zone, err)
			return
		}
	}
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
