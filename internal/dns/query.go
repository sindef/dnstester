package dns

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"dnstester/pkg/types"

	"github.com/miekg/dns"
)

// QueryDNS performs a DNS query using the specified protocol (udp, tcp, dot, doh).
// Uses github.com/miekg/dns for UDP/TCP/DoT and net/http for DoH. Extracts A and AAAA records.
// ResponseTime is measured in milliseconds.
func QueryDNS(server types.Server, domain string, protocol string) types.QueryResult {
	result := types.QueryResult{
		ServerName:    server.Name,
		ServerAddress: server.Address,
		Domain:        domain,
		Protocol:      protocol,
		ResponseIPs:   []string{},
		Success:       false,
	}

	startTime := time.Now()

	switch strings.ToLower(protocol) {
	case "udp":
		err := queryUDP(server.Address, domain, &result)
		if err != nil {
			result.Error = err.Error()
		}
	case "tcp":
		err := queryTCP(server.Address, domain, &result)
		if err != nil {
			result.Error = err.Error()
		}
	case "dot":
		err := queryDoT(server.Address, domain, &result)
		if err != nil {
			result.Error = err.Error()
		}
	case "doh":
		err := queryDoH(server.Address, domain, &result)
		if err != nil {
			result.Error = err.Error()
		}
	default:
		result.Error = fmt.Sprintf("unsupported protocol: %s", protocol)
	}

	result.ResponseTime = time.Since(startTime).Milliseconds()
	return result
}

// queryUDP performs a DNS query over UDP (port 53). Uses github.com/miekg/dns.
// Defaults to port 53 if no port is specified in the address.
func queryUDP(address string, domain string, result *types.QueryResult) error {
	addr := address
	if !strings.Contains(addr, ":") {
		addr = net.JoinHostPort(addr, "53")
	}

	client := &dns.Client{
		Net:     "udp",
		Timeout: 10 * time.Second,
	}

	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	r, _, err := client.Exchange(msg, addr)
	if err != nil {
		return err
	}

	if r.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("DNS query failed with RCODE: %d", r.Rcode)
	}

	for _, answer := range r.Answer {
		if a, ok := answer.(*dns.A); ok {
			result.ResponseIPs = append(result.ResponseIPs, a.A.String())
		}
		if aaaa, ok := answer.(*dns.AAAA); ok {
			result.ResponseIPs = append(result.ResponseIPs, aaaa.AAAA.String())
		}
	}

	result.Success = true
	return nil
}

// queryTCP performs a DNS query over TCP (port 53). Uses github.com/miekg/dns.
// Defaults to port 53 if no port is specified in the address.
func queryTCP(address string, domain string, result *types.QueryResult) error {
	addr := address
	if !strings.Contains(addr, ":") {
		addr = net.JoinHostPort(addr, "53")
	}

	client := &dns.Client{
		Net:     "tcp",
		Timeout: 10 * time.Second,
	}

	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	r, _, err := client.Exchange(msg, addr)
	if err != nil {
		return err
	}

	if r.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("DNS query failed with RCODE: %d", r.Rcode)
	}

	for _, answer := range r.Answer {
		if a, ok := answer.(*dns.A); ok {
			result.ResponseIPs = append(result.ResponseIPs, a.A.String())
		}
		if aaaa, ok := answer.(*dns.AAAA); ok {
			result.ResponseIPs = append(result.ResponseIPs, aaaa.AAAA.String())
		}
	}

	result.Success = true
	return nil
}

// queryDoT performs a DNS query over DNS-over-TLS (port 853). Uses github.com/miekg/dns with tcp-tls.
// Defaults to port 853 if no port is specified. TLS ServerName is extracted from the address hostname.
func queryDoT(address string, domain string, result *types.QueryResult) error {
	addr := address
	if !strings.Contains(addr, ":") {
		addr = net.JoinHostPort(addr, "853")
	}

	client := &dns.Client{
		Net:       "tcp-tls",
		TLSConfig: &tls.Config{ServerName: strings.Split(addr, ":")[0]},
		Timeout:   10 * time.Second,
	}

	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	r, _, err := client.Exchange(msg, addr)
	if err != nil {
		return err
	}

	if r.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("DNS query failed with RCODE: %d", r.Rcode)
	}

	for _, answer := range r.Answer {
		if a, ok := answer.(*dns.A); ok {
			result.ResponseIPs = append(result.ResponseIPs, a.A.String())
		}
		if aaaa, ok := answer.(*dns.AAAA); ok {
			result.ResponseIPs = append(result.ResponseIPs, aaaa.AAAA.String())
		}
	}

	result.Success = true
	return nil
}

// queryDoH performs a DNS query over DNS-over-HTTPS using net/http.
// Automatically constructs the DoH URL: adds https:// prefix if missing and appends /dns-query if needed.
// Sends DNS message as binary POST with Content-Type: application/dns-message.
func queryDoH(address string, domain string, result *types.QueryResult) error {
	url := address
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	if !strings.Contains(url, "/dns-query") && !strings.Contains(url, "/resolve") {
		if !strings.HasSuffix(url, "/") {
			url += "/"
		}
		url += "dns-query"
	}

	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	buf, err := msg.Pack()
	if err != nil {
		return fmt.Errorf("failed to pack DNS message: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	respBuf := make([]byte, 4096)
	n, err := resp.Body.Read(respBuf)
	if err != nil && err.Error() != "EOF" {
		return fmt.Errorf("failed to read response: %w", err)
	}

	response := new(dns.Msg)
	if err := response.Unpack(respBuf[:n]); err != nil {
		return fmt.Errorf("failed to unpack DNS response: %w", err)
	}

	if response.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("DNS query failed with RCODE: %d", response.Rcode)
	}

	for _, answer := range response.Answer {
		if a, ok := answer.(*dns.A); ok {
			result.ResponseIPs = append(result.ResponseIPs, a.A.String())
		}
		if aaaa, ok := answer.(*dns.AAAA); ok {
			result.ResponseIPs = append(result.ResponseIPs, aaaa.AAAA.String())
		}
	}

	result.Success = true
	return nil
}
