package db

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/logger"
)

// Config holds the database configuration
type Config struct {
	Host                string     `json:"host"`
	Port                int        `json:"port"`
	Keyspace            string     `json:"keyspace"`
	Username            string     `json:"username"`
	Password            string     `json:"password"`
	RequireConfirmation bool       `json:"requireConfirmation,omitempty"`
	SSL                 *SSLConfig `json:"ssl,omitempty"`
	AI                  *AIConfig  `json:"ai,omitempty"`
}

// SSLConfig holds SSL/TLS configuration options
type SSLConfig struct {
	Enabled            bool   `json:"enabled"`
	CertPath           string `json:"certPath,omitempty"`           // Path to client certificate
	KeyPath            string `json:"keyPath,omitempty"`            // Path to client private key
	CAPath             string `json:"caPath,omitempty"`             // Path to CA certificate
	HostVerification   bool   `json:"hostVerification,omitempty"`   // Enable hostname verification
	InsecureSkipVerify bool   `json:"insecureSkipVerify,omitempty"` // Skip certificate verification (not recommended for production)
}

// AIConfig holds AI provider configuration
type AIConfig struct {
	Provider  string            `json:"provider"` // "mock", "openai", "anthropic", "gemini", "ollama"
	APIKey    string            `json:"apiKey"`   // General API key (overridden by provider-specific)
	Model     string            `json:"model"`    // General model (overridden by provider-specific)
	OpenAI    *AIProviderConfig `json:"openai,omitempty"`
	Anthropic *AIProviderConfig `json:"anthropic,omitempty"`
	Gemini    *AIProviderConfig `json:"gemini,omitempty"`
	Ollama    *AIProviderConfig `json:"ollama,omitempty"`
}

// AIProviderConfig holds provider-specific configuration
type AIProviderConfig struct {
	APIKey string `json:"apiKey"`
	Model  string `json:"model"`
	URL    string `json:"url,omitempty"` // For local providers like Ollama
}

// OutputFormat represents the output format for query results
type OutputFormat string

const (
	OutputFormatTable  OutputFormat = "TABLE"
	OutputFormatASCII  OutputFormat = "ASCII"
	OutputFormatExpand OutputFormat = "EXPAND"
	OutputFormatJSON   OutputFormat = "JSON"
)

// Session is a wrapper around the gocql.Session.
type Session struct {
	*gocql.Session
	cluster             *gocql.ClusterConfig
	consistency         gocql.Consistency
	pageSize            int
	tracing             bool
	currentKeyspace     string
	requireConfirmation bool
	cassandraVersion    string
	aiConfig            *AIConfig
	outputFormat        OutputFormat
	schemaCache         *SchemaCache
}

// SessionOptions represents options for creating a session with command-line overrides
type SessionOptions struct {
	Host                string
	Port                int
	Keyspace            string
	Username            string
	Password            string
	RequireConfirmation bool
	SSL                 *SSLConfig
}

// NewSession creates a new Cassandra session.
func NewSession() (*Session, error) {
	return NewSessionWithOptions(SessionOptions{})
}

// customLogger suppresses gocql error messages to prevent terminal corruption
type customLogger struct{}

func (c *customLogger) Error(msg string, fields ...gocql.LogField)   {}
func (c *customLogger) Warning(msg string, fields ...gocql.LogField) {}
func (c *customLogger) Info(msg string, fields ...gocql.LogField)    {}
func (c *customLogger) Debug(msg string, fields ...gocql.LogField)   {}

// NewSessionWithOptions creates a new Cassandra session with command-line overrides.
func NewSessionWithOptions(options SessionOptions) (*Session, error) {
	// Also redirect standard log output to discard
	log.SetOutput(io.Discard)

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		// Use defaults if config file not found
		config = &Config{
			Host:                "127.0.0.1",
			Port:                9042,
			Keyspace:            "",
			Username:            "cassandra",
			Password:            "cassandra",
			RequireConfirmation: true,
			AI: &AIConfig{
				Provider: "mock",
			},
		}
	}

	// Override config with command-line options if provided
	if options.Host != "" {
		config.Host = options.Host
	}
	if options.Port != 0 {
		config.Port = options.Port
	}
	if options.Keyspace != "" {
		config.Keyspace = options.Keyspace
	}
	if options.Username != "" {
		config.Username = options.Username
	}
	if options.Password != "" {
		config.Password = options.Password
	}
	// RequireConfirmation is handled specially since false is a valid value
	// Only override if explicitly set via command line
	if options.Host != "" || options.Port != 0 || options.Keyspace != "" ||
		options.Username != "" || options.Password != "" {
		// If any command-line option was provided, use the RequireConfirmation from options
		config.RequireConfirmation = options.RequireConfirmation
	}

	// Override SSL config if provided
	if options.SSL != nil {
		config.SSL = options.SSL
	}

	// Create cluster configuration
	cluster := gocql.NewCluster(fmt.Sprintf("%s:%d", config.Host, config.Port))
	// Suppress gocql's default logging to prevent terminal corruption
	cluster.Logger = &customLogger{}
	cluster.Consistency = gocql.LocalOne
	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 10 * time.Second
	cluster.ProtoVersion = 4
	cluster.DisableInitialHostLookup = true

	if config.Keyspace != "" {
		cluster.Keyspace = config.Keyspace
	}

	if config.Username != "" && config.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: config.Username,
			Password: config.Password,
		}
	}

	// Configure SSL if enabled
	if config.SSL != nil && config.SSL.Enabled {
		tlsConfig, err := createTLSConfig(config.SSL)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS configuration: %v", err)
		}
		cluster.SslOpts = &gocql.SslOptions{
			Config: tlsConfig,
		}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cassandra: %v", err)
	}

	// Get Cassandra version
	var releaseVersion string
	iter := session.Query("SELECT release_version FROM system.local").Iter()
	iter.Scan(&releaseVersion)
	iter.Close()

	s := &Session{
		Session:             session,
		cluster:             cluster,
		consistency:         gocql.LocalOne,
		pageSize:            100,
		tracing:             false,
		currentKeyspace:     config.Keyspace,
		requireConfirmation: config.RequireConfirmation,
		cassandraVersion:    releaseVersion,
		aiConfig:            config.AI,
		outputFormat:        OutputFormatTable,
	}

	// Initialize schema cache for AI features
	s.schemaCache = NewSchemaCache(s)
	if err := s.schemaCache.Refresh(); err != nil {
		// Log error but don't fail connection - AI features will work without cache
		logger.DebugfToFile("Session", "Failed to initialize schema cache: %v", err)
	} else {
		logger.DebugfToFile("Session", "Schema cache initialized with %d keyspaces", len(s.schemaCache.Keyspaces))
	}

	return s, nil
}

// loadConfig loads the configuration from cqlai.json
func loadConfig() (*Config, error) {
	// Look for config file in current directory
	configPath := "cqlai.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Try home directory
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ".cqlai.json")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Ensure AI config exists with default values
	if config.AI == nil {
		config.AI = &AIConfig{
			Provider: "mock",
		}
	}

	overrideWithEnvVars(&config)
	return &config, nil
}

// overrideWithEnvVars overrides the configuration with values from environment variables.
func overrideWithEnvVars(config *Config) {
	if host := os.Getenv("CASSANDRA_HOST"); host != "" {
		config.Host = host
	}
	if portStr := os.Getenv("CASSANDRA_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Port = port
		}
	}
	if keyspace := os.Getenv("CASSANDRA_KEYSPACE"); keyspace != "" {
		config.Keyspace = keyspace
	}
	if user := os.Getenv("CASSANDRA_USER"); user != "" {
		config.Username = user
	}
	if pass := os.Getenv("CASSANDRA_PASSWORD"); pass != "" {
		config.Password = pass
	}

	if config.AI == nil {
		config.AI = &AIConfig{}
	}

	if provider := os.Getenv("AI_PROVIDER"); provider != "" {
		config.AI.Provider = provider
	}
	if model := os.Getenv("AI_MODEL"); model != "" {
		config.AI.Model = model
	}

	// Ollama specific
	if config.AI.Ollama == nil {
		config.AI.Ollama = &AIProviderConfig{}
	}
	if url := os.Getenv("OLLAMA_URL"); url != "" {
		config.AI.Ollama.URL = url
	}
	if key := os.Getenv("OLLAMA_API_KEY"); key != "" {
		config.AI.Ollama.APIKey = key
	}
	if model := os.Getenv("OLLAMA_MODEL"); model != "" {
		config.AI.Ollama.Model = model
	}

	// OpenAI specific
	if config.AI.OpenAI == nil {
		config.AI.OpenAI = &AIProviderConfig{}
	}
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		config.AI.OpenAI.APIKey = key
	}
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		config.AI.OpenAI.Model = model
	}

	// Anthropic specific
	if config.AI.Anthropic == nil {
		config.AI.Anthropic = &AIProviderConfig{}
	}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		config.AI.Anthropic.APIKey = key
	}
	if model := os.Getenv("ANTHROPIC_MODEL"); model != "" {
		config.AI.Anthropic.Model = model
	}

	// Gemini specific
	if config.AI.Gemini == nil {
		config.AI.Gemini = &AIProviderConfig{}
	}
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		config.AI.Gemini.APIKey = key
	}
	if model := os.Getenv("GEMINI_MODEL"); model != "" {
		config.AI.Gemini.Model = model
	}
}

// Consistency returns the current consistency level
func (s *Session) Consistency() string {
	switch s.consistency {
	case gocql.Any:
		return "ANY"
	case gocql.One:
		return "ONE"
	case gocql.Two:
		return "TWO"
	case gocql.Three:
		return "THREE"
	case gocql.Quorum:
		return "QUORUM"
	case gocql.All:
		return "ALL"
	case gocql.LocalQuorum:
		return "LOCAL_QUORUM"
	case gocql.EachQuorum:
		return "EACH_QUORUM"
	case gocql.LocalOne:
		return "LOCAL_ONE"
	default:
		return "UNKNOWN"
	}
}

// SetConsistency sets the consistency level
func (s *Session) SetConsistency(level string) error {
	var consistency gocql.Consistency
	switch level {
	case "ANY":
		consistency = gocql.Any
	case "ONE":
		consistency = gocql.One
	case "TWO":
		consistency = gocql.Two
	case "THREE":
		consistency = gocql.Three
	case "QUORUM":
		consistency = gocql.Quorum
	case "ALL":
		consistency = gocql.All
	case "LOCAL_QUORUM":
		consistency = gocql.LocalQuorum
	case "EACH_QUORUM":
		consistency = gocql.EachQuorum
	case "LOCAL_ONE":
		consistency = gocql.LocalOne
	default:
		return fmt.Errorf("invalid consistency level: %s", level)
	}
	s.consistency = consistency
	return nil
}

// PageSize returns the current page size
func (s *Session) PageSize() int {
	return s.pageSize
}

// SetPageSize sets the page size
func (s *Session) SetPageSize(size int) {
	s.pageSize = size
}

// Tracing returns whether tracing is enabled
func (s *Session) Tracing() bool {
	return s.tracing
}

// SetTracing enables or disables tracing
func (s *Session) SetTracing(enabled bool) {
	s.tracing = enabled
}

// GetOutputFormat returns the current output format
func (s *Session) GetOutputFormat() OutputFormat {
	return s.outputFormat
}

// SetOutputFormat sets the output format
func (s *Session) SetOutputFormat(format string) error {
	logger.DebugfToFile("SetOutputFormat", "Setting format to: %s", format)
	switch strings.ToUpper(format) {
	case "TABLE":
		s.outputFormat = OutputFormatTable
		logger.DebugfToFile("SetOutputFormat", "Format set to TABLE: %v", s.outputFormat)
	case "ASCII":
		s.outputFormat = OutputFormatASCII
		logger.DebugfToFile("SetOutputFormat", "Format set to ASCII: %v", s.outputFormat)
	case "EXPAND":
		s.outputFormat = OutputFormatExpand
		logger.DebugfToFile("SetOutputFormat", "Format set to EXPAND: %v", s.outputFormat)
	case "JSON":
		s.outputFormat = OutputFormatJSON
		logger.DebugfToFile("SetOutputFormat", "Format set to JSON: %v", s.outputFormat)
	default:
		return fmt.Errorf("invalid output format '%s'. Use OUTPUT TABLE, OUTPUT ASCII, OUTPUT EXPAND, or OUTPUT JSON", format)
	}
	return nil
}

// Query creates a new query with session defaults applied
func (s *Session) Query(stmt string, values ...interface{}) *gocql.Query {
	// If we have a current keyspace and the query is a SELECT/INSERT/UPDATE/DELETE
	// without an explicit keyspace, prepend the keyspace
	processedStmt := s.prependKeyspaceIfNeeded(stmt)

	query := s.Session.Query(processedStmt, values...)
	query.Consistency(s.consistency)
	query.PageSize(s.pageSize)
	// Note: Tracing would be enabled per query with query.Observer()
	// For now, we'll handle tracing display separately
	return query
}

// prependKeyspaceIfNeeded adds the current keyspace to table references if needed
func (s *Session) prependKeyspaceIfNeeded(stmt string) string {
	// This is a simplified implementation
	// In production, you'd want a proper CQL parser
	if s.currentKeyspace == "" {
		return stmt
	}

	// Don't modify system queries or queries that already have keyspace
	upperStmt := strings.ToUpper(strings.TrimSpace(stmt))
	if strings.Contains(upperStmt, "SYSTEM") || strings.Contains(stmt, ".") {
		return stmt
	}

	// Only modify SELECT/INSERT/UPDATE/DELETE/TRUNCATE statements
	if strings.HasPrefix(upperStmt, "SELECT") ||
		strings.HasPrefix(upperStmt, "INSERT") ||
		strings.HasPrefix(upperStmt, "UPDATE") ||
		strings.HasPrefix(upperStmt, "DELETE") ||
		strings.HasPrefix(upperStmt, "TRUNCATE") {
		// This is a very basic implementation
		// A real implementation would need proper CQL parsing
		return stmt
	}

	return stmt
}

// CurrentKeyspace returns the current keyspace
func (s *Session) CurrentKeyspace() string {
	return s.currentKeyspace
}

// SetKeyspace sets the current keyspace
func (s *Session) SetKeyspace(keyspace string) {
	s.currentKeyspace = keyspace
}

// RequireConfirmation returns whether confirmation is required for dangerous commands
func (s *Session) RequireConfirmation() bool {
	return s.requireConfirmation
}

// SetRequireConfirmation sets whether confirmation is required for dangerous commands
func (s *Session) SetRequireConfirmation(require bool) {
	s.requireConfirmation = require
}

// CassandraVersion returns the Cassandra version
func (s *Session) CassandraVersion() string {
	if s.cassandraVersion == "" {
		return "unknown"
	}
	return s.cassandraVersion
}

// IsVersion4OrHigher checks if the Cassandra version is 4.0 or higher
func (s *Session) IsVersion4OrHigher() bool {
	version := s.CassandraVersion()
	// Parse version string like "4.0.4" or "5.0.4"
	parts := strings.Split(version, ".")
	if len(parts) < 1 {
		return false
	}

	majorVersion, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}

	return majorVersion >= 4
}

// GetAIConfig returns the AI configuration
func (s *Session) GetAIConfig() *AIConfig {
	return s.aiConfig
}

// GetSchemaCache returns the schema cache
func (s *Session) GetSchemaCache() *SchemaCache {
	return s.schemaCache
}

// createTLSConfig creates a TLS configuration based on the SSL settings
func createTLSConfig(sslConfig *SSLConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: sslConfig.InsecureSkipVerify,
	}

	// Load client certificate if provided
	if sslConfig.CertPath != "" && sslConfig.KeyPath != "" {
		cert, err := tls.LoadX509KeyPair(sslConfig.CertPath, sslConfig.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if provided
	if sslConfig.CAPath != "" {
		caCert, err := os.ReadFile(sslConfig.CAPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %v", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Configure hostname verification
	if !sslConfig.HostVerification {
		tlsConfig.InsecureSkipVerify = true
	}

	return tlsConfig, nil
}
