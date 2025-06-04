// SPDX-FileCopyrightText: (c) 2016 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package gcloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"slices"
	"time"

	"github.com/czerwonk/dns-drain/pkg/changelog"
	"github.com/czerwonk/dns-drain/pkg/drain"

	dns "google.golang.org/api/dns/v1"
)

type GoogleDnsDrainer struct {
	cfg     Config
	service *dns.Service
	logger  changelog.ChangeLogger
	updater *recordUpdater
	opt     *drain.Options
}

type DrainFilter func(*dns.ResourceRecordSet) []string

func NewDrainer(cfg Config, logger changelog.ChangeLogger, opt *drain.Options) *GoogleDnsDrainer {
	return &GoogleDnsDrainer{
		cfg:    cfg,
		logger: logger,
		opt:    opt,
	}
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

func (client *GoogleDnsDrainer) DrainWithRegex(regex *regexp.Regexp, newValue string) error {
	filter := func(rec *dns.ResourceRecordSet) []string {
		return filterWithRegex(rec, regex)
	}

	return client.performForZones(filter, newValue)
}

func (client *GoogleDnsDrainer) performForZones(filter DrainFilter, newValue string) error {
	ctx := context.Background()
	svc, err := dns.NewService(ctx, client.cfg.toClientOptions()...)
	if err != nil {
		return err
	}
	client.service = svc

	client.updater = &recordUpdater{
		service: client.service,
		project: client.cfg.Project,
		dryRun:  client.opt.DryRun,
		limit:   client.opt.Limit,
	}

	zones, err := client.getZones()
	if err != nil {
		return err
	}

	doneCh := make(chan bool)
	defer close(doneCh)

	for _, z := range zones {
		go client.drainForZone(z.Name, filter, newValue, doneCh)
	}

	for range zones {
		select {
		case <-doneCh:
		case <-time.After(2 * time.Minute):
			return errors.New(fmt.Sprintln("Timeout exceeded!"))
		}
	}

	return nil
}

func (client *GoogleDnsDrainer) getZones() ([]*dns.ManagedZone, error) {
	r, err := client.service.ManagedZones.List(client.cfg.Project).Do()
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
	return client.opt.SkipFilter != nil && client.opt.SkipFilter.MatchString(zone)
}

func (client *GoogleDnsDrainer) matchesZoneFilter(zone string) bool {
	return client.opt.ZoneFilter == nil || client.opt.ZoneFilter.MatchString(zone)
}

func (client *GoogleDnsDrainer) drainForZone(zone string, filter DrainFilter, newValue string, doneCh chan bool) {
	defer func() { doneCh <- true }()

	r, err := client.service.ResourceRecordSets.List(client.cfg.Project, zone).Do()
	if err != nil {
		log.Printf("ERROR - %s: %s\n", zone, err)
		return
	}

	for _, rec := range r.Rrsets {
		if !client.matchesNameFilter(rec.Name) {
			continue
		}

		client.handleRecordSet(zone, rec, newValue, filter)
	}
}

func (client *GoogleDnsDrainer) matchesNameFilter(name string) bool {
	return client.opt.NameFilter == nil || client.opt.NameFilter.MatchString(name)
}

func (client *GoogleDnsDrainer) handleRecordSet(zone string, rec *dns.ResourceRecordSet, newValue string, filter DrainFilter) {
	if len(client.opt.TypeFilter) > 0 && client.opt.TypeFilter != rec.Type {
		return
	}

	d := filter(rec)

	if len(d) == 0 && len(newValue) == 0 && !client.opt.Force {
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

func filterWithRegex(rec *dns.ResourceRecordSet, regex *regexp.Regexp) []string {
	res := make([]string, 0)

	for _, x := range rec.Rrdatas {
		if !regex.MatchString(x) {
			res = append(res, x)
		}
	}

	return res
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
	return slices.Contains(datas, value)
}

func (client *GoogleDnsDrainer) updateRecordSet(rec *dns.ResourceRecordSet, zone string, datas []string) error {
	done, err := client.updater.updateRecordSet(zone, rec, datas)
	if err != nil {
		return err
	}

	if done {
		return client.logChanges(rec, zone, datas)
	}

	return nil
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
