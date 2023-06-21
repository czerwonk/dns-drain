// SPDX-FileCopyrightText: (c) 2016 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("dns-drainctl")
		fmt.Printf("Version: %s\n", version)
		fmt.Println("Author(s): Daniel Czerwonk")
		fmt.Println("Drain and undrain frontends by using DNS")
	},
}
