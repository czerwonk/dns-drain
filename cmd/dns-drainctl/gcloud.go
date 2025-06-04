// SPDX-FileCopyrightText: (c) 2016 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/czerwonk/dns-drain/pkg/changelog"
	"github.com/czerwonk/dns-drain/pkg/drain"
	"github.com/czerwonk/dns-drain/pkg/gcloud"
	"github.com/czerwonk/dns-drain/pkg/undrain"
	"github.com/spf13/cobra"
)

type gcloudCommand struct{}

var gcloudCmd = &cobra.Command{
	Use:     "google-cloud",
	Aliases: []string{"gcloud"},
	Short:   "Drain and undrain DNS records using Google Cloud API",
}

func init() {
	g := &gcloudCommand{}

	gcloudCmd.PersistentFlags().String("project", "", "Name of the Google Cloud project")
	gcloudCmd.PersistentFlags().String("credentials-file", "", "Path to the cloud credentials file (if not set, cloud SDK will be used)")
	addDrainCommand(gcloudCmd, g.drainer)
	addUndrainCommand(gcloudCmd, g.undrainer)
}

func (g *gcloudCommand) drainer(cmd *cobra.Command, logger changelog.ChangeLogger, opt *drain.Options) drain.Drainer {
	cfg := configFromArgs()
	return gcloud.NewDrainer(cfg, logger, opt)
}

func (g *gcloudCommand) undrainer(cmd *cobra.Command, opt *undrain.Options) undrain.Undrainer {
	cfg := configFromArgs()
	return gcloud.NewUndrainer(cfg, opt)
}

func configFromArgs() gcloud.Config {
	project, _ := gcloudCmd.PersistentFlags().GetString("project")
	if project == "" {
		cobra.CheckErr(fmt.Errorf("please specify the Google Cloud project"))
	}

	credentialsFile, _ := gcloudCmd.PersistentFlags().GetString("credentials-file")
	return gcloud.Config{
		Project:         project,
		CredentialsFile: credentialsFile,
	}
}
