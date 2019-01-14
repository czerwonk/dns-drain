package gcloud

import (
	"log"
	"reflect"

	"sync/atomic"

	dns "google.golang.org/api/dns/v1"
)

type recordUpdater struct {
	service *dns.Service
	project string
	dryRun  bool
	limit   int64
	counter int64
}

func (u *recordUpdater) updateRecordSet(zone string, rec *dns.ResourceRecordSet, datas []string) (bool, error) {
	if reflect.DeepEqual(rec.Rrdatas, datas) {
		return false, nil
	}

	count := atomic.AddInt64(&u.counter, 1)
	if u.limit >= 0 && count > u.limit {
		return false, nil
	}

	if len(rec.Rrdatas) > 0 {
		log.Printf("- %s: %s %s\n", rec.Name, rec.Type, rec.Rrdatas)
	}

	if len(datas) > 0 {
		log.Printf("+ %s: %s %s\n", rec.Name, rec.Type, datas)
	}

	if u.dryRun {
		return true, nil
	}

	c := &dns.Change{Additions: make([]*dns.ResourceRecordSet, 0), Deletions: make([]*dns.ResourceRecordSet, 0)}
	if len(rec.Rrdatas) > 0 {
		c.Deletions = append(c.Deletions, rec)
	}

	if len(datas) > 0 {
		updated := *rec
		updated.Rrdatas = datas
		c.Additions = append(c.Additions, &updated)
	}

	_, err := u.service.Changes.Create(u.project, zone, c).Do()
	if err != nil {
		return false, err
	}

	return true, nil
}
