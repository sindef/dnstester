package types

// Config represents the root configuration structure
type Config struct {
	Domains  []string  `yaml:"domains"`
	Servers []Server `yaml:"servers"`
}

// Server represents a DNS server configuration
type Server struct {
	Name      string   `yaml:"name"`
	Address   string   `yaml:"address"`
	Protocols []string `yaml:"protocols"`
}

// QueryResult represents the result of a DNS query
type QueryResult struct {
	ServerName    string
	ServerAddress string
	Domain        string
	Protocol      string
	ResponseIPs   []string
	ResponseTime  int64 // milliseconds
	Success       bool
	Error         string
}

// Report represents the complete test report
type Report struct {
	Results []QueryResult
	Summary Summary
}

// Summary contains aggregate statistics
type Summary struct {
	TotalQueries   int
	Successful     int
	Failed         int
	AverageTime    float64
	MinTime        int64
	MaxTime        int64
}

