package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("dns-drainctl")
		fmt.Printf("Version: %s\n", version)
		fmt.Println("Author(s): Daniel Czerwonk")
		fmt.Println("Drain and undrain frontends by using DNS")
	},
}
