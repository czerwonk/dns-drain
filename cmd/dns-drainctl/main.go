package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "1.0.0"

var (
	rootCmd = &cobra.Command{
		Use:   "dns-drainctl",
		Short: "Drain by removing/replacing IP/net from DNS records with ease",
		Example: `
Drain IP 1.2.3.4 in project api-project-xxx by removing IP from records
$ dns-drainctl gcloud --project api-project-xxx drain -f drain.json 1.2.3.4/32

Drain IP 1.2.3.4 in project api-project-xxx by replacing IP with 1.2.3.5
$ dns-drainctl gcloud --project api-project-xxx drain 1.2.3.4/32 -f drain.json --replace-by 1.2.3.5

Undrain by using json file written in drain process
$ dns-drainctl gcloud --project api-project-xxx undrain -f drain.json`,
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(gcloudCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
