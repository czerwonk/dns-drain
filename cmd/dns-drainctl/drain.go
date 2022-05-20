package main

import (
	"fmt"
	"log"
	"net"
	"regexp"

	"github.com/czerwonk/dns-drain/pkg/changelog"
	"github.com/czerwonk/dns-drain/pkg/drain"
	"github.com/spf13/cobra"
)

type DrainerFunc func(*cobra.Command, changelog.ChangeLogger, *drain.Options) drain.Drainer

func addDrainCommand(cmd *cobra.Command, d DrainerFunc) {
	drainCmd := &cobra.Command{
		Use:   "drain",
		Short: "Removes or replaces DNS records",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			performDrainCommand(cmd, args, d)
		},
	}
	drainCmd.PersistentFlags().Bool("dry", false, "Do not modify DNS records (simulation only)")
	drainCmd.PersistentFlags().StringP("file", "f", "drain.json", "Changelog file")
	drainCmd.PersistentFlags().StringP("zone", "z", "", "Apply only to zones matching the specified regex")
	drainCmd.PersistentFlags().String("skip", "", "Skip zones matching the specified regex")
	drainCmd.PersistentFlags().String("type", "", "Record type to change")
	drainCmd.PersistentFlags().Int64("limit", -1, "Max number of records to change (-1 = unlimited)")
	drainCmd.PersistentFlags().Bool("force", false, "Remove value from record even if it is the only value")
	drainCmd.PersistentFlags().Bool("use-regex", false, "Regex to find data in DNS records to remove/replace")
	drainCmd.PersistentFlags().String("replace-by", "", "Value to replace the matched data by (empty = no replacement)")

	cmd.AddCommand(drainCmd)
}

func performDrainCommand(cmd *cobra.Command, args []string, d DrainerFunc) {
	f, _ := cmd.PersistentFlags().GetString("file")
	if len(f) == 0 {
		cobra.CheckErr(fmt.Errorf("please provide a path for the changelog"))
	}

	logger, err := changelog.NewFileChangeLogger(f)
	cobra.CheckErr(err)
	defer flushAndCloseLogger(logger)

	opt := optionsFromDrainCommand(cmd)
	drainer := d(cmd, logger, opt)

	if opt.DryRun {
		log.Println("Using dry run. No records will be changed.")
	}

	if opt.Force {
		log.Println("Logic check was disabled. There is no guarantee for a consistent result.")
	}

	pattern := args[0]
	replacement, _ := cmd.PersistentFlags().GetString("replace-by")
	useRegex, _ := cmd.PersistentFlags().GetBool("use-regex")

	err = performDrain(pattern, replacement, useRegex, drainer)
	cobra.CheckErr(err)
}

func optionsFromDrainCommand(cmd *cobra.Command) *drain.Options {
	opt := &drain.Options{}

	opt.DryRun, _ = cmd.PersistentFlags().GetBool("dry")
	opt.Force, _ = cmd.PersistentFlags().GetBool("force")
	opt.Limit, _ = cmd.PersistentFlags().GetInt64("limit")
	opt.TypeFilter, _ = cmd.PersistentFlags().GetString("type")

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

func flushAndCloseLogger(logger *changelog.FileChangeLogger) {
	err := logger.Flush()
	cobra.CheckErr(err)

	err = logger.Close()
	cobra.CheckErr(err)
}

func performDrain(pattern, replacement string, useRegex bool, d drain.Drainer) error {
	if useRegex {
		return performDrainWithRegex(pattern, replacement, d)
	}

	if ipNet, found := extractIPNetwork(pattern); found {
		return performDrainWithIPNetwork(ipNet, replacement, d)
	}

	return d.DrainWithValue(pattern, replacement)
}

func performDrainWithRegex(pattern, replacement string, d drain.Drainer) error {
	r, err := regexp.Compile(pattern)
	if err != nil {
		cobra.CheckErr(fmt.Errorf("invalid regex pattern: %w", err))
	}

	return d.DrainWithRegex(r, replacement)
}

func performDrainWithIPNetwork(ipNet *net.IPNet, replacement string, d drain.Drainer) error {
	if len(replacement) == 0 {
		return d.DrainWithIpNet(ipNet, nil)
	}

	replaceIP := net.ParseIP(replacement)
	if replaceIP == nil {
		cobra.CheckErr(fmt.Errorf("please specify valid IP for replacement when using IP as matcher"))
	}

	return d.DrainWithIpNet(ipNet, replaceIP)
}

func extractIPNetwork(s string) (*net.IPNet, bool) {
	_, ipNet, err := net.ParseCIDR(s)

	if err != nil {
		ipAddr := net.ParseIP(s)
		if ipAddr == nil {
			return nil, false
		}

		if ipAddr.To4() != nil {
			_, ipNet, _ = net.ParseCIDR(fmt.Sprintf("%s/32", ipAddr))
		} else {
			_, ipNet, _ = net.ParseCIDR(fmt.Sprintf("%s/128", ipAddr))
		}
	}

	return ipNet, true
}
