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

// perform a query uing the specified protocol
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

func queryUDP(address string, domain string, result *types.QueryResult) error {
	// Ensure address has port
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

func queryDoT(address string, domain string, result *types.QueryResult) error {
	// For DoT, we need to use port 853 by default - can be overridden by user
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

func queryDoH(address string, domain string, result *types.QueryResult) error {
	// For DoH, we need to construct the proper URL
	// DoH endpoints typically look like: https://dns.server/dns-query
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
