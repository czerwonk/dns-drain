package gcloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/oauth2/google"

	dns "google.golang.org/api/dns/v1"
)

type GoogleDnsClient struct {
	Project string
	DryRun  bool
}

func NewClient(project string, dryRun bool) *GoogleDnsClient {
	return &GoogleDnsClient{Project: project, DryRun: dryRun}
}

func (client *GoogleDnsClient) Drain(ipNet *net.IPNet) error {
	ctx := context.Background()
	c, err := google.DefaultClient(ctx, dns.CloudPlatformScope)
	if err != nil {
		return err
	}

	d, err := dns.New(c)
	if err != nil {
		return err
	}

	r, err := d.ManagedZones.List(client.Project).Do()
	if err != nil {
		return err
	}

	done := make(chan bool)
	for _, z := range r.ManagedZones {
		go client.drainWithIpNet(d, z.Name, ipNet, done)
	}

	for i := 0; i < len(r.ManagedZones); i++ {
		select {
		case <-done:
		case <-time.After(60 * time.Second):
			return errors.New(fmt.Sprintln("Timeout exceeded!"))
		}
	}

	return nil
}

func (client *GoogleDnsClient) drainWithIpNet(s *dns.Service, zone string, ipNet *net.IPNet, done chan bool) {
	defer func() { done <- true }()

	r, err := s.ResourceRecordSets.List(client.Project, zone).Do()
	if err != nil {
		log.Printf("ERROR - %s: %s\n", zone, err)
		return
	}

	for _, rec := range r.Rrsets {
		d := filterWithIpNet(rec, ipNet)

		if len(d) > 0 {
			if len(d) != len(rec.Rrdatas) {
				err := client.updateRecordSet(s, rec, zone, d)
				if err != nil {
					log.Printf("ERROR - %s: %s", rec.Name, err)
				}
			}
		} else {
			log.Printf("WARN - %s %s: Only one IP assigned to record. Can not drain!\n", rec.Type, rec.Name)
		}
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

func (client *GoogleDnsClient) updateRecordSet(s *dns.Service, rec *dns.ResourceRecordSet, zone string, datas []string) error {
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

	_, err := s.Changes.Create(client.Project, zone, c).Do()
	return err
}
