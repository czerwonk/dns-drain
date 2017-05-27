package main

import (
	"flag"
	"fmt"
)

func printUsage() {
	fmt.Println("Usage: dns-drain [ ... ]\n\nParameters:")
	fmt.Println()
	flag.PrintDefaults()
	fmt.Println("\n\nExamples:")
	fmt.Println()
	fmt.Println("Drain IP 1.2.3.4 in project api-project-xxx by removing IP from records")
	fmt.Println("  dns-drain -gcloud.project=api-project-xxx -ip=1.2.3.4/32")
	fmt.Println()
	fmt.Println("Drain IP 1.2.3.4 in project api-project-xxx by replacing IP with 1.2.3.5")
	fmt.Println("  dns-drain -gcloud.project=api-project-xxx -ip=1.2.3.4/32 -new_ip=1.2.3.5")
	fmt.Println()
	fmt.Println("Undrain by using json file written in drain process")
	fmt.Println("  dns-drain -undrain -gcloud.project=api-project-xxx")
}
