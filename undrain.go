package main

import (
	"github.com/czerwonk/dns-drain/changelog"
	"github.com/czerwonk/dns-drain/gcloud"
)

func undrain(f string) error {
	changeLog := changelog.NewFileChangeLog(f)
	c, err := changeLog.GetChanges()
	if err != nil {
		return err
	}

	u := gcloud.NewUndrainer(*gcloudProject, *dry, *zoneFilter)
	return u.Undrain(c)
}