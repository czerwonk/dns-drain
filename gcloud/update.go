package gcloud

import (
	"log"
	"reflect"

	dns "google.golang.org/api/dns/v1"
)

type updateContext struct {
	service *dns.Service
	project string
	zone    string
	dryRun  bool
}

func updateRecordSet(rec *dns.ResourceRecordSet, datas []string, context *updateContext) error {
	if reflect.DeepEqual(rec.Rrdatas, datas) {
		return nil
	}

	log.Printf("- %s: %s %s\n", rec.Name, rec.Type, rec.Rrdatas)
	log.Printf("+ %s: %s %s\n", rec.Name, rec.Type, datas)

	if context.dryRun {
		return nil
	}

	c := &dns.Change{Additions: make([]*dns.ResourceRecordSet, 0), Deletions: make([]*dns.ResourceRecordSet, 0)}
	c.Deletions = append(c.Deletions, rec)

	updated := *rec
	updated.Rrdatas = datas
	c.Additions = append(c.Additions, &updated)

	_, err := context.service.Changes.Create(context.project, context.zone, c).Do()
	return err
}
