package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"dnstester/internal/dns"
	"dnstester/internal/report"
	"dnstester/pkg/types"
)

// StartServer starts the HTTP server
func StartServer(addr string) error {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/test", handleTest)
	http.HandleFunc("/api/report", handleReport)

	log.Printf("Server listening on http://localhost%s", addr)
	return http.ListenAndServe(addr, nil)
}

// handleIndex serves the WebUI
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DNS Tester</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 30px;
        }
        h1 {
            color: #333;
            margin-bottom: 30px;
            text-align: center;
        }
        .form-section {
            margin-bottom: 30px;
        }
        .form-section h2 {
            color: #555;
            font-size: 1.2em;
            margin-bottom: 15px;
            border-bottom: 2px solid #667eea;
            padding-bottom: 8px;
        }
        .input-group {
            margin-bottom: 15px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            color: #666;
            font-weight: 500;
        }
        input[type="text"], textarea {
            width: 100%;
            padding: 10px;
            border: 2px solid #ddd;
            border-radius: 6px;
            font-size: 14px;
            transition: border-color 0.3s;
        }
        input[type="text"]:focus, textarea:focus {
            outline: none;
            border-color: #667eea;
        }
        textarea {
            resize: vertical;
            min-height: 80px;
            font-family: monospace;
        }
        .server-item {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 6px;
            margin-bottom: 15px;
            border: 1px solid #e0e0e0;
        }
        .server-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
        }
        .server-name {
            font-weight: 600;
            color: #333;
        }
        .btn-remove {
            background: #dc3545;
            color: white;
            border: none;
            padding: 5px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 12px;
        }
        .btn-remove:hover {
            background: #c82333;
        }
        .protocols {
            display: flex;
            gap: 15px;
            flex-wrap: wrap;
            margin-top: 10px;
        }
        .protocol-checkbox {
            display: flex;
            align-items: center;
            gap: 5px;
        }
        .protocol-checkbox input[type="checkbox"] {
            width: auto;
        }
        .btn {
            background: #667eea;
            color: white;
            border: none;
            padding: 12px 30px;
            border-radius: 6px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: background 0.3s;
            margin-right: 10px;
        }
        .btn:hover {
            background: #5568d3;
        }
        .btn-secondary {
            background: #6c757d;
        }
        .btn-secondary:hover {
            background: #5a6268;
        }
        .btn:disabled {
            background: #ccc;
            cursor: not-allowed;
        }
        .loading {
            display: none;
            text-align: center;
            padding: 20px;
            color: #667eea;
        }
        .loading.active {
            display: block;
        }
        .spinner {
            border: 3px solid #f3f3f3;
            border-top: 3px solid #667eea;
            border-radius: 50%;
            width: 40px;
            height: 40px;
            animation: spin 1s linear infinite;
            margin: 0 auto 10px;
        }
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        #results {
            margin-top: 30px;
            display: none;
        }
        #results.active {
            display: block;
        }
        .summary {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 6px;
            margin-bottom: 20px;
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 15px;
        }
        .summary-item {
            text-align: center;
        }
        .summary-label {
            font-size: 12px;
            color: #666;
            text-transform: uppercase;
            margin-bottom: 5px;
        }
        .summary-value {
            font-size: 24px;
            font-weight: 600;
            color: #333;
        }
        .results-table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
        }
        .results-table th {
            background: #667eea;
            color: white;
            padding: 12px;
            text-align: left;
            font-weight: 600;
        }
        .results-table td {
            padding: 12px;
            border-bottom: 1px solid #e0e0e0;
        }
        .results-table tr:hover {
            background: #f8f9fa;
        }
        .status-success {
            color: #28a745;
            font-weight: 600;
        }
        .status-failed {
            color: #dc3545;
            font-weight: 600;
        }
        .error-message {
            background: #f8d7da;
            color: #721c24;
            padding: 15px;
            border-radius: 6px;
            margin-top: 20px;
            border: 1px solid #f5c6cb;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌐 DNS Tester</h1>
        
        <form id="testForm">
            <div class="form-section">
                <h2>Domains to Test</h2>
                <div class="input-group">
                    <label for="domains">Enter domains (one per line):</label>
                    <textarea id="domains" name="domains" placeholder="github.com&#10;google.com&#10;cloudflare.com" required></textarea>
                </div>
            </div>

            <div class="form-section">
                <h2>DNS Servers</h2>
                <div id="servers">
                    <div class="server-item">
                        <div class="server-header">
                            <span class="server-name">Server 1</span>
                            <button type="button" class="btn-remove" onclick="removeServer(this)">Remove</button>
                        </div>
                        <div class="input-group">
                            <label>Server Name:</label>
                            <input type="text" name="server-name" placeholder="e.g., Cloudflare DNS" required>
                        </div>
                        <div class="input-group">
                            <label>Server Address:</label>
                            <input type="text" name="server-address" placeholder="e.g., 1.1.1.1 or dns.server.com" required>
                        </div>
                        <div class="protocols">
                            <div class="protocol-checkbox">
                                <input type="checkbox" name="protocol" value="udp" id="protocol-udp-0" checked>
                                <label for="protocol-udp-0">UDP</label>
                            </div>
                            <div class="protocol-checkbox">
                                <input type="checkbox" name="protocol" value="tcp" id="protocol-tcp-0">
                                <label for="protocol-tcp-0">TCP</label>
                            </div>
                            <div class="protocol-checkbox">
                                <input type="checkbox" name="protocol" value="dot" id="protocol-dot-0">
                                <label for="protocol-dot-0">DoT</label>
                            </div>
                            <div class="protocol-checkbox">
                                <input type="checkbox" name="protocol" value="doh" id="protocol-doh-0">
                                <label for="protocol-doh-0">DoH</label>
                            </div>
                        </div>
                    </div>
                </div>
                <button type="button" class="btn btn-secondary" onclick="addServer()">+ Add Server</button>
            </div>

            <div class="loading" id="loading">
                <div class="spinner"></div>
                <p>Running DNS tests... This may take a moment.</p>
            </div>

            <div style="margin-top: 20px;">
                <button type="submit" class="btn" id="submitBtn">Run Tests</button>
                <button type="button" class="btn btn-secondary" onclick="resetForm()">Reset</button>
            </div>
        </form>

        <div id="results">
            <h2 style="margin-top: 30px; margin-bottom: 20px;">Test Results</h2>
            <div id="summary" class="summary"></div>
            <div id="resultsTable"></div>
        </div>
    </div>

    <script>
        let serverIdCounter = 0;

        function getNextServerId() {
            return 'server-' + (serverIdCounter++);
        }

        function updateServerNumbers() {
            const serversDiv = document.getElementById('servers');
            const serverItems = serversDiv.querySelectorAll('.server-item');
            serverItems.forEach((item, index) => {
                const serverNameSpan = item.querySelector('.server-name');
                if (serverNameSpan) {
                    serverNameSpan.textContent = 'Server ' + (index + 1);
                }
            });
        }

        function addServer() {
            const serversDiv = document.getElementById('servers');
            const newServer = document.createElement('div');
            newServer.className = 'server-item';
            const uniqueId = getNextServerId();
            const serverNumber = serversDiv.children.length + 1;
            newServer.innerHTML = 
                '<div class="server-header">' +
                    '<span class="server-name">Server ' + serverNumber + '</span>' +
                    '<button type="button" class="btn-remove" onclick="removeServer(this)">Remove</button>' +
                '</div>' +
                '<div class="input-group">' +
                    '<label>Server Name:</label>' +
                    '<input type="text" name="server-name" placeholder="e.g., Cloudflare DNS" required>' +
                '</div>' +
                '<div class="input-group">' +
                    '<label>Server Address:</label>' +
                    '<input type="text" name="server-address" placeholder="e.g., 1.1.1.1 or dns.server.com" required>' +
                '</div>' +
                '<div class="protocols">' +
                    '<div class="protocol-checkbox">' +
                        '<input type="checkbox" name="protocol" value="udp" id="protocol-udp-' + uniqueId + '" checked>' +
                        '<label for="protocol-udp-' + uniqueId + '">UDP</label>' +
                    '</div>' +
                    '<div class="protocol-checkbox">' +
                        '<input type="checkbox" name="protocol" value="tcp" id="protocol-tcp-' + uniqueId + '">' +
                        '<label for="protocol-tcp-' + uniqueId + '">TCP</label>' +
                    '</div>' +
                    '<div class="protocol-checkbox">' +
                        '<input type="checkbox" name="protocol" value="dot" id="protocol-dot-' + uniqueId + '">' +
                        '<label for="protocol-dot-' + uniqueId + '">DoT</label>' +
                    '</div>' +
                    '<div class="protocol-checkbox">' +
                        '<input type="checkbox" name="protocol" value="doh" id="protocol-doh-' + uniqueId + '">' +
                        '<label for="protocol-doh-' + uniqueId + '">DoH</label>' +
                    '</div>' +
                '</div>';
            serversDiv.appendChild(newServer);
        }

        function removeServer(btn) {
            const serversDiv = document.getElementById('servers');
            if (serversDiv.children.length > 1) {
                btn.closest('.server-item').remove();
                updateServerNumbers();
            } else {
                alert('At least one server is required');
            }
        }

        function resetForm() {
            document.getElementById('testForm').reset();
            document.getElementById('results').classList.remove('active');
            document.getElementById('results').style.display = 'none';
            const serversDiv = document.getElementById('servers');
            while (serversDiv.children.length > 1) {
                serversDiv.removeChild(serversDiv.lastChild);
            }
            updateServerNumbers();
        }

        document.getElementById('testForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const loading = document.getElementById('loading');
            const submitBtn = document.getElementById('submitBtn');
            const results = document.getElementById('results');
            
            loading.classList.add('active');
            submitBtn.disabled = true;
            results.classList.remove('active');
            results.style.display = 'none';

            // Collect domains
            const domainsText = document.getElementById('domains').value;
            const domains = domainsText.split('\n')
                .map(d => d.trim())
                .filter(d => d.length > 0);

            // Collect servers
            const serverItems = document.querySelectorAll('.server-item');
            const servers = [];
            
            serverItems.forEach((item, index) => {
                const name = item.querySelector('input[name="server-name"]').value.trim();
                const address = item.querySelector('input[name="server-address"]').value.trim();
                const protocolCheckboxes = item.querySelectorAll('input[name="protocol"]:checked');
                const protocols = Array.from(protocolCheckboxes).map(cb => cb.value);
                
                if (name && address && protocols.length > 0) {
                    servers.push({
                        name: name,
                        address: address,
                        protocols: protocols
                    });
                }
            });

            if (domains.length === 0) {
                alert('Please enter at least one domain');
                loading.classList.remove('active');
                submitBtn.disabled = false;
                return;
            }

            if (servers.length === 0) {
                alert('Please configure at least one server with at least one protocol');
                loading.classList.remove('active');
                submitBtn.disabled = false;
                return;
            }

            try {
                const response = await fetch('/api/test', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        domains: domains,
                        servers: servers
                    })
                });

                if (!response.ok) {
                    throw new Error('Test failed: ' + response.statusText);
                }

                const data = await response.json();
                displayResults(data);
            } catch (error) {
                alert('Error running tests: ' + error.message);
            } finally {
                loading.classList.remove('active');
                submitBtn.disabled = false;
            }
        });

        function displayResults(data) {
            const results = document.getElementById('results');
            const summary = document.getElementById('summary');
            const resultsTable = document.getElementById('resultsTable');

            // Display summary
            const avgTime = data.summary.average_time ? data.summary.average_time.toFixed(2) : '0.00';
            summary.innerHTML = 
                '<div class="summary-item">' +
                    '<div class="summary-label">Total Queries</div>' +
                    '<div class="summary-value">' + data.summary.total_queries + '</div>' +
                '</div>' +
                '<div class="summary-item">' +
                    '<div class="summary-label">Successful</div>' +
                    '<div class="summary-value" style="color: #28a745;">' + data.summary.successful + '</div>' +
                '</div>' +
                '<div class="summary-item">' +
                    '<div class="summary-label">Failed</div>' +
                    '<div class="summary-value" style="color: #dc3545;">' + data.summary.failed + '</div>' +
                '</div>' +
                '<div class="summary-item">' +
                    '<div class="summary-label">Avg Time</div>' +
                    '<div class="summary-value">' + avgTime + ' ms</div>' +
                '</div>' +
                '<div class="summary-item">' +
                    '<div class="summary-label">Min Time</div>' +
                    '<div class="summary-value">' + (data.summary.min_time || 0) + ' ms</div>' +
                '</div>' +
                '<div class="summary-item">' +
                    '<div class="summary-label">Max Time</div>' +
                    '<div class="summary-value">' + (data.summary.max_time || 0) + ' ms</div>' +
                '</div>';

            // Display results table
            let tableHTML = '<table class="results-table"><thead><tr><th>Server</th><th>Address</th><th>Domain</th><th>Protocol</th><th>Response IPs</th><th>Time (ms)</th><th>Status</th><th>Error</th></tr></thead><tbody>';
            
            data.results.forEach(result => {
                const status = result.success ? 
                    '<span class="status-success">✓ Success</span>' : 
                    '<span class="status-failed">✗ Failed</span>';
                const ips = result.response_ips && result.response_ips.length > 0 ? 
                    result.response_ips.join(', ') : 'N/A';
                const error = result.error || '-';
                
                tableHTML += 
                    '<tr>' +
                        '<td>' + escapeHtml(result.server_name) + '</td>' +
                        '<td>' + escapeHtml(result.server_address) + '</td>' +
                        '<td>' + escapeHtml(result.domain) + '</td>' +
                        '<td>' + escapeHtml(result.protocol.toUpperCase()) + '</td>' +
                        '<td>' + escapeHtml(ips) + '</td>' +
                        '<td>' + result.response_time + '</td>' +
                        '<td>' + status + '</td>' +
                        '<td>' + escapeHtml(error) + '</td>' +
                    '</tr>';
            });
            
            tableHTML += '</tbody></table>';
            resultsTable.innerHTML = tableHTML;

            results.classList.add('active');
            results.style.display = 'block';
            results.scrollIntoView({ behavior: 'smooth' });
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
    </script>
</body>
</html>`

	t, err := template.New("index").Parse(tmpl)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.Execute(w, nil); err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		return
	}
}

// TestRequest represents the request body for running tests
type TestRequest struct {
	Domains []string      `json:"domains"`
	Servers []types.Server `json:"servers"`
}

// handleTest handles the test API endpoint
func handleTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.Domains) == 0 {
		http.Error(w, "At least one domain is required", http.StatusBadRequest)
		return
	}
	if len(req.Servers) == 0 {
		http.Error(w, "At least one server is required", http.StatusBadRequest)
		return
	}

	// Run tests
	var results []types.QueryResult
	for _, server := range req.Servers {
		for _, domain := range req.Domains {
			for _, protocol := range server.Protocols {
				result := dns.QueryDNS(server, domain, protocol)
				results = append(results, result)
			}
		}
	}

	// Generate summary
	summary := report.CalculateSummary(results)

	// Prepare response
	response := map[string]interface{}{
		"results": convertResults(results),
		"summary": convertSummary(summary),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleReport handles report generation (for future use)
func handleReport(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// convertResults converts QueryResult to JSON-friendly format
func convertResults(results []types.QueryResult) []map[string]interface{} {
	converted := make([]map[string]interface{}, len(results))
	for i, r := range results {
		converted[i] = map[string]interface{}{
			"server_name":    r.ServerName,
			"server_address": r.ServerAddress,
			"domain":         r.Domain,
			"protocol":       r.Protocol,
			"response_ips":   r.ResponseIPs,
			"response_time":  r.ResponseTime,
			"success":        r.Success,
			"error":          r.Error,
		}
	}
	return converted
}

// convertSummary converts Summary to JSON-friendly format
func convertSummary(summary types.Summary) map[string]interface{} {
	return map[string]interface{}{
		"total_queries": summary.TotalQueries,
		"successful":    summary.Successful,
		"failed":        summary.Failed,
		"average_time":  summary.AverageTime,
		"min_time":      summary.MinTime,
		"max_time":      summary.MaxTime,
	}
}

