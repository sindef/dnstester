package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"dnstester/internal/config"
	"dnstester/internal/dns"
	"dnstester/internal/report"
	"dnstester/internal/server"
	"dnstester/pkg/types"
)

func main() {
	var configFile string
	var outputFile string
	var csvOutput bool
	var serverMode bool
	var serverAddr string

	flag.StringVar(&configFile, "config", "config.yaml", "Path to YAML configuration file")
	flag.StringVar(&outputFile, "output", "", "Path to output report file (default: stdout)")
	flag.BoolVar(&csvOutput, "csv", false, "Output report in CSV format")
	flag.BoolVar(&serverMode, "server", false, "Run in server mode (HTTP WebUI)")
	flag.StringVar(&serverAddr, "addr", ":8080", "Server address (default: :8080)")
	flag.Parse()

	// If server mode, start HTTP server
	if serverMode {
		log.Printf("Starting DNS Tester server on %s", serverAddr)
		if err := server.StartServer(serverAddr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
		return
	}

	// Original CLI mode
	if configFile == "" {
		fmt.Fprintf(os.Stderr, "Error: config file is required\n")
		flag.Usage()
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var results []types.QueryResult

	fmt.Println("Starting DNS tests...")
	fmt.Printf("Testing %d domain(s) against %d server(s)...\n\n", len(cfg.Domains), len(cfg.Servers))

	for _, server := range cfg.Servers {
		fmt.Printf("Testing server: %s (%s)\n", server.Name, server.Address)

		for _, domain := range cfg.Domains {
			for _, protocol := range server.Protocols {
				fmt.Printf("  Querying %s via %s...\n", domain, protocol)
				result := dns.QueryDNS(server, domain, protocol)
				results = append(results, result)

				if result.Success {
					fmt.Printf("    ✓ Success: %s (Time: %d ms)\n",
						formatIPs(result.ResponseIPs), result.ResponseTime)
				} else {
					fmt.Printf("    ✗ Failed: %s\n", result.Error)
				}
			}
		}
		fmt.Println()
	}

	// Generate report
	fmt.Println("Generating report...")
	if err := report.GenerateReport(results, outputFile, csvOutput); err != nil {
		log.Fatalf("Failed to generate report: %v", err)
	}

	if outputFile != "" {
		fmt.Printf("\nReport saved to: %s\n", outputFile)
	}
}

// formatIPs formats a slice of IP addresses for display
func formatIPs(ips []string) string {
	if len(ips) == 0 {
		return "No IPs"
	}
	if len(ips) == 1 {
		return ips[0]
	}
	return fmt.Sprintf("%d IPs: %s", len(ips), strings.Join(ips, ", "))
}
