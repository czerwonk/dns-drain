// SPDX-FileCopyrightText: (c) 2016 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"log"
	"regexp"

	"github.com/czerwonk/dns-drain/pkg/changelog"
	"github.com/czerwonk/dns-drain/pkg/undrain"
	"github.com/spf13/cobra"
)

type UndrainerFunc func(*cobra.Command, *undrain.Options) undrain.Undrainer

func addUndrainCommand(cmd *cobra.Command, u UndrainerFunc) {
	undrainCmd := &cobra.Command{
		Use:   "undrain",
		Short: "Rollback DNS changes by using the changelog file",
		Run: func(cmd *cobra.Command, args []string) {
			performUndrainCommand(cmd, args, u)
		},
	}
	undrainCmd.PersistentFlags().Bool("dry", false, "Do not modify DNS records (simulation only)")
	undrainCmd.PersistentFlags().StringP("file", "f", "drain.json", "File containing changes to revert")
	undrainCmd.PersistentFlags().StringP("zone", "z", "", "Apply only to zones matching the specified regex")
	undrainCmd.PersistentFlags().String("skip", "", "Skip zones matching the specified regex")
	undrainCmd.PersistentFlags().Int64("limit", -1, "Max number of records to change (-1 = unlimited)")

	cmd.AddCommand(undrainCmd)
}

func performUndrainCommand(cmd *cobra.Command, args []string, u UndrainerFunc) {
	f, _ := cmd.PersistentFlags().GetString("file")
	if len(f) == 0 {
		cobra.CheckErr(fmt.Errorf("please provide a path for the changelog source file"))
	}

	changeLog := changelog.NewFileChangeLog(f)
	c, err := changeLog.GetChanges()
	cobra.CheckErr(err)

	opt := optionsFromUndrainCommand(cmd)
	undrainer := u(cmd, opt)

	if opt.DryRun {
		log.Println("Using dry run. No records will be changed.")
	}

	err = undrainer.Undrain(c)
	cobra.CheckErr(err)
}

func optionsFromUndrainCommand(cmd *cobra.Command) *undrain.Options {
	opt := &undrain.Options{}

	opt.DryRun, _ = cmd.PersistentFlags().GetBool("dry")
	opt.Limit, _ = cmd.PersistentFlags().GetInt64("limit")

	zoneFilter, _ := cmd.PersistentFlags().GetString("zone")
	if len(zoneFilter) > 0 {
		r, err := regexp.Compile(zoneFilter)
		if err != nil {
			cobra.CheckErr(fmt.Errorf("invalid zone filter regex: %w", err))
		}
		opt.ZoneFilter = r
	}

	skipFilter, _ := cmd.PersistentFlags().GetString("skip")
	if len(skipFilter) > 0 {
		r, err := regexp.Compile(skipFilter)
		if err != nil {
			cobra.CheckErr(fmt.Errorf("invalid skip filter regex: %w", err))
		}
		opt.SkipFilter = r
	}

	return opt
}
