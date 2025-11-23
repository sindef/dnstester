package report

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"dnstester/pkg/types"
)

// GenerateReport generates a formatted report. Uses text/tabwriter for text format and encoding/csv for CSV.
// If outputFile is empty, writes to stdout. CSV format uses semicolon-separated IPs.
func GenerateReport(results []types.QueryResult, outputFile string, csvFormat bool) error {
	report := &types.Report{
		Results: results,
		Summary: CalculateSummary(results),
	}

	// Create output file or use stdout if not specificed
	var writer *os.File
	var err error
	if outputFile != "" {
		writer, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer writer.Close()
	} else {
		writer = os.Stdout
	}

	if csvFormat {
		return generateCSVReport(writer, report)
	}

	// Write report header
	fmt.Fprintf(writer, "DNS Tester Report\n")
	fmt.Fprintf(writer, "==================\n\n")

	writeSummary(writer, report.Summary)

	fmt.Fprintf(writer, "\nDetailed Results\n")
	fmt.Fprintf(writer, "================\n\n")

	tw := tabwriter.NewWriter(writer, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "Server\tAddress\tDomain\tProtocol\tResponse IPs\tTime (ms)\tStatus\tError")
	fmt.Fprintln(tw, "------\t-------\t------\t--------\t------------\t---------\t------\t-----")

	for _, result := range results {
		status := "✓"
		if !result.Success {
			status = "✗"
		}

		ips := strings.Join(result.ResponseIPs, ", ")
		if ips == "" {
			ips = "N/A"
		}

		errorMsg := result.Error
		if errorMsg == "" {
			errorMsg = "-"
		}

		// Ensure response time is always shown, even for failed queries
		responseTime := result.ResponseTime
		if responseTime < 0 {
			responseTime = 0
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
			result.ServerName,
			result.ServerAddress,
			result.Domain,
			result.Protocol,
			ips,
			responseTime,
			status,
			errorMsg,
		)
	}

	tw.Flush()

	return nil
}

// generateCSVReport writes a CSV report using encoding/csv. Response IPs are semicolon-separated.
func generateCSVReport(writer *os.File, report *types.Report) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	header := []string{"Server", "Address", "Domain", "Protocol", "Response IPs", "Time (ms)", "Status", "Error"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, result := range report.Results {
		status := "Success"
		if !result.Success {
			status = "Failed"
		}

		ips := strings.Join(result.ResponseIPs, "; ")
		if ips == "" {
			ips = "N/A"
		}

		errorMsg := result.Error
		if errorMsg == "" {
			errorMsg = ""
		}

		responseTime := result.ResponseTime
		if responseTime < 0 {
			responseTime = 0
		}

		row := []string{
			result.ServerName,
			result.ServerAddress,
			result.Domain,
			result.Protocol,
			ips,
			fmt.Sprintf("%d", responseTime),
			status,
			errorMsg,
		}

		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// CalculateSummary calculates aggregate statistics. Only successful queries are included in timing calculations.
// MinTime is initialized to -1 to distinguish "no successful queries" from "min time is 0ms".
func CalculateSummary(results []types.QueryResult) types.Summary {
	summary := types.Summary{
		TotalQueries: len(results),
		MinTime:      -1,
		MaxTime:      0,
	}

	var totalTime int64
	var successfulCount int

	for _, result := range results {
		if result.Success {
			summary.Successful++
			successfulCount++
			totalTime += result.ResponseTime

			if summary.MinTime == -1 || result.ResponseTime < summary.MinTime {
				summary.MinTime = result.ResponseTime
			}
			if result.ResponseTime > summary.MaxTime {
				summary.MaxTime = result.ResponseTime
			}
		} else {
			summary.Failed++
		}
	}

	if successfulCount > 0 {
		summary.AverageTime = float64(totalTime) / float64(successfulCount)
	}

	return summary
}

func writeSummary(writer *os.File, summary types.Summary) {
	fmt.Fprintf(writer, "Summary\n")
	fmt.Fprintf(writer, "-------\n")
	fmt.Fprintf(writer, "Total Queries:    %d\n", summary.TotalQueries)
	fmt.Fprintf(writer, "Successful:       %d\n", summary.Successful)
	fmt.Fprintf(writer, "Failed:           %d\n", summary.Failed)
	if summary.Successful > 0 {
		fmt.Fprintf(writer, "Average Time:     %.2f ms\n", summary.AverageTime)
		fmt.Fprintf(writer, "Min Time:         %d ms\n", summary.MinTime)
		fmt.Fprintf(writer, "Max Time:         %d ms\n", summary.MaxTime)
	}
}
