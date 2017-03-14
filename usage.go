package main

import (
	"flag"
	"fmt"
)

func printUsage() {
	fmt.Println("Usage: dns-drain [ ... ]\n\nParameters:\n")
	flag.PrintDefaults()
	fmt.Println("\n\nExamples:\n")
	fmt.Println("Drain IP 1.2.3.4 in project api-project-xxx by removing IP from records\n")
	fmt.Println("  dns-drain -gcloud.project=api-project-xxx -ip=1.2.3.4/32\n")
	fmt.Println("Drain IP 1.2.3.4 in project api-project-xxx by replacing IP with 1.2.3.5\n")
	fmt.Println("  dns-drain -gcloud.project=api-project-xxx -ip=1.2.3.4/32 -new_ip=1.2.3.5\n")
	fmt.Println("Undrain by using json file written in drain process\n")
	fmt.Println("  dns-drain -undrain -gcloud.project=api-project-xxx")
}
