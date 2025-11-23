# DNS Tester

A comprehensive DNS testing tool written in Go that tests DNS resolution across multiple servers, domains, and protocols (UDP, TCP, DoT, DoH).

## Features

- Test DNS queries against multiple servers
- Support for multiple protocols:
  - **UDP** (port 53)
  - **TCP** (port 53)
  - **DoT** (DNS-over-TLS, port 853)
  - **DoH** (DNS-over-HTTPS)
- Test all domains against all servers (global domain list)
- Generate detailed reports with response times, IPs, and success/failure status
- Output reports in text or CSV format
- YAML-based configuration
- **WebUI server mode** - Interactive web interface for running tests

## Project Structure

```
dnstester/
├── cmd/
│   └── dnstester/
│       └── main.go          # Main entry point
├── internal/
│   ├── config/
│   │   └── config.go        # YAML config parser
│   ├── dns/
│   │   └── query.go         # DNS query implementations
│   ├── report/
│   │   └── report.go        # Report generation
│   └── server/
│       └── server.go        # HTTP server and WebUI
├── pkg/
│   └── types/
│       └── types.go         # Shared types
├── config.yaml              # Example configuration file
├── go.mod                   # Go module dependencies
└── README.md                # This file
```

## Installation

1. Clone or navigate to the project directory
2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Building

Build the project:
```bash
go build -o dnstester ./cmd/dnstester
```


## Usage

1. Create a configuration file (see `config.yaml` for reference)
2. Run the tester:
   ```bash
   ./dnstester -config config.yaml
   ```

3. Save report to file:
   ```bash
   ./dnstester -config config.yaml -output report.txt
   ```

4. Generate CSV report:
   ```bash
   ./dnstester -config config.yaml -output report.csv -csv
   ```

5. Run in server mode (WebUI):
   ```bash
   ./dnstester -server
   ```
   Then open your browser to `http://localhost:8080`

### Command Line Options

- `-config`: Path to YAML configuration file (default: `config.yaml`)
- `-output`: Path to output report file (default: stdout)
- `-csv`: Output report in CSV format instead of text format
- `-server`: Run in server mode (HTTP WebUI)
- `-addr`: Server address when in server mode (default: `:8080`)

## Configuration File Format

The configuration file is a YAML file with the following structure. Note that domains are defined globally and will be tested against all servers:

```yaml
domains:
  - "example.com"
  - "google.com"
  - "github.com"

servers:
  - name: "Server Name"
    address: "dns.server.ip"
    protocols:
      - "udp"
      - "tcp"
      - "dot"
      - "doh"
```

### Protocol Specifications

- **udp**: Standard DNS over UDP (port 53)
- **tcp**: Standard DNS over TCP (port 53)
- **dot**: DNS-over-TLS (port 853). Address should be the server IP or hostname
- **doh**: DNS-over-HTTPS. Address should be the full URL (e.g., `https://cloudflare-dns.com/dns-query`)

### Example Configuration

See `config.yaml` for a complete example with multiple servers and protocols.

## Report Format

The generated report can be output in two formats:

### Text Format (default)

1. **Summary Section**:
   - Total queries
   - Successful queries
   - Failed queries
   - Average response time
   - Min/Max response times

2. **Detailed Results**:
   - Server name and address
   - Domain tested
   - Protocol used
   - Response IP addresses
   - Response time (milliseconds)
   - Success/failure status
   - Error messages (if any)

### CSV Format

When using the `-csv` flag, the report is generated as a CSV file with the following columns:
- Server
- Address
- Domain
- Protocol
- Response IPs (semicolon-separated)
- Time (ms)
- Status
- Error

## Server Mode (WebUI)

The DNS Tester includes a web-based user interface that allows you to run tests interactively without needing a configuration file.

### Starting the Server

```bash
./dnstester -server
```

By default, the server listens on `:8080`. You can specify a different address:

```bash
./dnstester -server -addr :9090
```

### Using the WebUI

1. Open your browser and navigate to `http://localhost:8080` (or your custom address)
2. Enter the domains you want to test (one per line)
3. Configure DNS servers:
   - Click "+ Add Server" to add more servers
   - For each server, provide:
     - Server name (e.g., "Cloudflare DNS")
     - Server address (e.g., "1.1.1.1" or "dns.server.com")
     - Select protocols to test (UDP, TCP, DoT, DoH)
4. Click "Run Tests" to execute the tests
5. View the results in the interactive report with:
   - Summary statistics (total queries, success/failure counts, timing metrics)
   - Detailed results table with all query information

The WebUI provides a modern, responsive interface that makes it easy to test DNS configurations on the fly without editing configuration files.

## Dependencies

- `github.com/miekg/dns` - DNS library for protocol support
- `gopkg.in/yaml.v3` - YAML parsing



