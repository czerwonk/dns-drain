package gcloud

import (
	"log"
	"reflect"

	dns "google.golang.org/api/dns/v1"
)

type recordUpdater struct {
	service *dns.Service
	project string
	dryRun  bool
}

func (u *recordUpdater) updateRecordSet(zone string, rec *dns.ResourceRecordSet, datas []string) error {
	if reflect.DeepEqual(rec.Rrdatas, datas) {
		return nil
	}

	log.Printf("- %s: %s %s\n", rec.Name, rec.Type, rec.Rrdatas)
	log.Printf("+ %s: %s %s\n", rec.Name, rec.Type, datas)

	if u.dryRun {
		return nil
	}

	c := &dns.Change{Additions: make([]*dns.ResourceRecordSet, 0), Deletions: make([]*dns.ResourceRecordSet, 0)}
	c.Deletions = append(c.Deletions, rec)

	updated := *rec
	updated.Rrdatas = datas
	c.Additions = append(c.Additions, &updated)

	_, err := u.service.Changes.Create(u.project, zone, c).Do()
	return err
}
