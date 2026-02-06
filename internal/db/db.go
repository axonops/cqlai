package db

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/logger"
)

// Session is a wrapper around the gocql.Session.
type Session struct {
	*gocql.Session
	cluster          *gocql.ClusterConfig
	consistency      gocql.Consistency
	pageSize         int
	tracing          bool
	autoFetch        bool   // Auto-fetch all pages without scroll pauses
	username         string // Current connection username
	cassandraVersion string
	schemaCache      *SchemaCache
	udtRegistry      *UDTRegistry
	lastTraceID      []byte // Store the last trace ID for retrieval
}

// SessionOptions represents options for creating a session with command-line overrides
type SessionOptions struct {
	Host           string
	Port           int
	Keyspace       string
	Username       string
	Password       string
	Consistency    string // Default consistency level (e.g., "LOCAL_ONE", "QUORUM")
	SSL            *config.SSLConfig
	BatchMode      bool   // Skip schema caching for batch mode
	ConnectTimeout int    // Connection timeout in seconds (0 = use default)
	RequestTimeout int    // Request timeout in seconds (0 = use default)
	ConfigFile     string // Path to custom config file
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
	cfg, err := loadConfig(options.ConfigFile)
	if err != nil {
		logger.DebugfToFile("Session", "loadConfig() failed: %v", err)
		// Use defaults if config file not found
		cfg = &config.Config{
			Host:                "127.0.0.1",
			Port:                9042,
			Keyspace:            "",
			Username:            "cassandra",
			Password:            "cassandra",
			RequireConfirmation: true,
			AI: &config.AIConfig{
				Provider: "mock",
			},
		}
		logger.DebugfToFile("Session", "Using default config: host=%s, port=%d, username=%s", 
			cfg.Host, cfg.Port, cfg.Username)
	} else {
		logger.DebugfToFile("Session", "Loaded config: host=%s, port=%d, username=%s, keyspace=%s, hasPassword=%v", 
			cfg.Host, cfg.Port, cfg.Username, cfg.Keyspace, cfg.Password != "")
	}

	// Override config with command-line options if provided
	if options.Host != "" {
		cfg.Host = options.Host
		logger.DebugfToFile("Session", "Overriding host with command-line option: %s", options.Host)
	}
	if options.Port != 0 {
		cfg.Port = options.Port
		logger.DebugfToFile("Session", "Overriding port with command-line option: %d", options.Port)
	}
	if options.Keyspace != "" {
		cfg.Keyspace = options.Keyspace
		logger.DebugfToFile("Session", "Overriding keyspace with command-line option: %s", options.Keyspace)
	}
	if options.Username != "" {
		cfg.Username = options.Username
		logger.DebugfToFile("Session", "Overriding username with command-line option: %s", options.Username)
	}
	if options.Password != "" {
		cfg.Password = options.Password
		logger.DebugfToFile("Session", "Overriding password with command-line option")
	}
	// Override SSL config if provided
	if options.SSL != nil {
		cfg.SSL = options.SSL
		logger.DebugfToFile("Session", "Overriding SSL config with command-line option")
	}
	
	// Log final configuration being used
	logger.DebugfToFile("Session", "Final config for connection: host=%s:%d, username=%s, keyspace=%s, hasPassword=%v", 
		cfg.Host, cfg.Port, cfg.Username, cfg.Keyspace, cfg.Password != "")

	// Create cluster configuration
	cluster := gocql.NewCluster(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	// Suppress gocql's default logging to prevent terminal corruption
	cluster.Logger = &customLogger{}
	cluster.Consistency = gocql.LocalOne
	
	// Set timeouts based on options, config, or use defaults
	switch {
	case options.RequestTimeout > 0:
		cluster.Timeout = time.Duration(options.RequestTimeout) * time.Second
	case cfg.RequestTimeout > 0:
		cluster.Timeout = time.Duration(cfg.RequestTimeout) * time.Second
	default:
		cluster.Timeout = 10 * time.Second
	}
	
	switch {
	case options.ConnectTimeout > 0:
		cluster.ConnectTimeout = time.Duration(options.ConnectTimeout) * time.Second
	case cfg.ConnectTimeout > 0:
		cluster.ConnectTimeout = time.Duration(cfg.ConnectTimeout) * time.Second
	default:
		cluster.ConnectTimeout = 10 * time.Second
	}
	
	cluster.DisableInitialHostLookup = true

	if cfg.Keyspace != "" {
		cluster.Keyspace = cfg.Keyspace
	}

	if cfg.Username != "" && cfg.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	// Configure SSL if enabled
	if cfg.SSL != nil && cfg.SSL.Enabled {
		tlsConfig, err := createTLSConfig(cfg.SSL, cfg.Host)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS configuration: %v", err)
		}
		cluster.SslOpts = &gocql.SslOptions{
			Config: tlsConfig,
		}
	}

	// Try to connect with progressively lower protocol versions
	// Protocol v5: Cassandra 3.10+, 4.0+, 5.0+
	// Protocol v4: Cassandra 3.0+
	// Protocol v3: Cassandra 2.1+
	var session *gocql.Session
	protocolVersions := []int{5, 4, 3}
	
	for _, protoVer := range protocolVersions {
		cluster.ProtoVersion = protoVer
		session, err = cluster.CreateSession()
		if err == nil {
			// Successfully connected
			logger.DebugfToFile("Session", "Connected with protocol version %d", protoVer)
			break
		}
		// Log the failure and try next version
		logger.DebugfToFile("Session", "Failed to connect with protocol version %d: %v", protoVer, err)
	}
	
	if session == nil {
		return nil, fmt.Errorf("failed to connect to Cassandra with any supported protocol version: %v", err)
	}

	// Get Cassandra version
	var releaseVersion string
	iter := session.Query("SELECT release_version FROM system.local").Iter()
	iter.Scan(&releaseVersion)
	_ = iter.Close()

	// Determine initial consistency level
	initialConsistency := gocql.LocalOne
	if options.Consistency != "" {
		// Use consistency from options (comes from config)
		switch strings.ToUpper(options.Consistency) {
		case "ANY":
			initialConsistency = gocql.Any
		case "ONE":
			initialConsistency = gocql.One
		case "TWO":
			initialConsistency = gocql.Two
		case "THREE":
			initialConsistency = gocql.Three
		case "QUORUM":
			initialConsistency = gocql.Quorum
		case "ALL":
			initialConsistency = gocql.All
		case "LOCAL_QUORUM":
			initialConsistency = gocql.LocalQuorum
		case "EACH_QUORUM":
			initialConsistency = gocql.EachQuorum
		case "LOCAL_ONE":
			initialConsistency = gocql.LocalOne
		default:
			logger.DebugfToFile("Session", "Invalid consistency level '%s', defaulting to LOCAL_ONE", options.Consistency)
		}
	} else if cfg.Consistency != "" {
		// Use consistency from config
		switch strings.ToUpper(cfg.Consistency) {
		case "ANY":
			initialConsistency = gocql.Any
		case "ONE":
			initialConsistency = gocql.One
		case "TWO":
			initialConsistency = gocql.Two
		case "THREE":
			initialConsistency = gocql.Three
		case "QUORUM":
			initialConsistency = gocql.Quorum
		case "ALL":
			initialConsistency = gocql.All
		case "LOCAL_QUORUM":
			initialConsistency = gocql.LocalQuorum
		case "EACH_QUORUM":
			initialConsistency = gocql.EachQuorum
		case "LOCAL_ONE":
			initialConsistency = gocql.LocalOne
		default:
			logger.DebugfToFile("Session", "Invalid consistency level '%s' in config, defaulting to LOCAL_ONE", cfg.Consistency)
		}
	}

	s := &Session{
		Session:          session,
		cluster:          cluster,
		consistency:      initialConsistency,
		pageSize:         100,
		tracing:          false,
		username:         cfg.Username,
		cassandraVersion: releaseVersion,
	}

	// Initialize schema cache for AI features (skip in batch mode)
	if !options.BatchMode {
		s.schemaCache = NewSchemaCache(s)
		if err := s.schemaCache.Refresh(); err != nil {
			// Log error but don't fail connection - AI features will work without cache
			logger.DebugfToFile("Session", "Failed to initialize schema cache: %v", err)
		} else {
			logger.DebugfToFile("Session", "Schema cache initialized with %d keyspaces", len(s.schemaCache.Keyspaces))
		}
	} else {
		logger.DebugfToFile("Session", "Skipping schema cache initialization in batch mode")
	}

	return s, nil
}

// loadConfig loads the configuration from cqlshrc and cqlai.json files
func loadConfig(customConfigPath string) (*config.Config, error) {
	// Use the proper config.LoadConfig() which handles cqlshrc files
	conf, err := config.LoadConfig(customConfigPath)
	if err != nil {
		return nil, err
	}

	// Ensure AI config exists with default values
	if conf.AI == nil {
		conf.AI = &config.AIConfig{
			Provider: "mock",
		}
	}

	return conf, nil
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

// AutoFetch returns whether auto-fetch is enabled
func (s *Session) AutoFetch() bool {
	return s.autoFetch
}

// SetAutoFetch enables or disables auto-fetching all pages
func (s *Session) SetAutoFetch(enabled bool) {
	s.autoFetch = enabled
}

// Username returns the current connection username
func (s *Session) Username() string {
	return s.username
}

// Query creates a new query with session defaults applied
func (s *Session) Query(stmt string, values ...interface{}) *gocql.Query {
	query := s.Session.Query(stmt, values...)
	query.Consistency(s.consistency)
	// Only set page size if it's greater than 0
	// PageSize 0 means use server default (no client-side paging control)
	if s.pageSize > 0 {
		query.PageSize(s.pageSize)
	}
	// Tracing will be handled in ExecuteSelectQuery when needed
	return query
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

// IsVersion3OrHigher checks if the Cassandra version is 3.0 or higher
func (s *Session) IsVersion3OrHigher() bool {
	version := s.CassandraVersion()
	// Parse version string like "3.0.4" or "4.0.4"
	parts := strings.Split(version, ".")
	if len(parts) < 1 {
		return false
	}

	majorVersion, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}

	return majorVersion >= 3
}

// GetSchemaCache returns the schema cache
func (s *Session) GetSchemaCache() *SchemaCache {
	return s.schemaCache
}

// TraceInfo holds trace session summary information
type TraceInfo struct {
	Coordinator string
	Duration    int
}

// GetTraceData retrieves trace data for the last executed query
func (s *Session) GetTraceData() ([][]string, []string, *TraceInfo, error) {
	if s.lastTraceID == nil {
		return nil, nil, nil, fmt.Errorf("no trace data available")
	}

	// Query the system_traces.events table for trace events
	// Note: Always use LOCAL_ONE consistency for system_traces queries
	// because trace data may not be replicated to all nodes yet
	query := `SELECT event_id, activity, source, source_elapsed, thread
	          FROM system_traces.events
	          WHERE session_id = ?
	          ORDER BY event_id`

	// Use LOCAL_ONE consistency for trace queries regardless of session consistency
	iter := s.Session.Query(query, s.lastTraceID).Consistency(gocql.LocalOne).Iter()
	defer iter.Close()

	// Define headers
	headers := []string{"Event", "Activity", "Source", "Source Elapsed (μs)", "Thread"}

	// Collect results
	var results [][]string

	var eventID gocql.UUID
	var activity, source, thread string
	var sourceElapsed int

	for iter.Scan(&eventID, &activity, &source, &sourceElapsed, &thread) {
		row := []string{
			eventID.String()[:8], // Short event ID
			activity,
			source,
			fmt.Sprintf("%d", sourceElapsed),
			thread,
		}
		results = append(results, row)
	}

	if err := iter.Close(); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to retrieve trace data: %v", err)
	}

	// Get session info - also use LOCAL_ONE consistency
	var traceInfo *TraceInfo
	var coordinator string
	var duration int
	sessionIter := s.Session.Query(`SELECT coordinator, duration
	                                FROM system_traces.sessions
	                                WHERE session_id = ?`, s.lastTraceID).Consistency(gocql.LocalOne).Iter()
	if sessionIter.Scan(&coordinator, &duration) {
		traceInfo = &TraceInfo{
			Coordinator: coordinator,
			Duration:    duration,
		}
	}
	_ = sessionIter.Close()

	return results, headers, traceInfo, nil
}

// Keyspace returns the current keyspace
func (s *Session) Keyspace() string {
	if s.cluster != nil {
		return s.cluster.Keyspace
	}
	return ""
}

// GetUDTRegistry returns the UDT registry
func (s *Session) GetUDTRegistry() *UDTRegistry {
	return s.udtRegistry
}

// SetUDTRegistry sets the UDT registry
func (s *Session) SetUDTRegistry(registry *UDTRegistry) {
	s.udtRegistry = registry
}

// GetColumnTypeFromSystemTable gets the full type definition for a column
// This method uses the metadata API when possible, falling back to system tables
func (s *Session) GetColumnTypeFromSystemTable(keyspace, table, column string) string {
	return s.getColumnTypeUsingMetadata(keyspace, table, column)
}

// SetKeyspace changes the current keyspace by recreating the session
func (s *Session) SetKeyspace(keyspace string) error {
	// Close the current session
	s.Close()

	// Update cluster config with new keyspace
	s.cluster.Keyspace = keyspace

	// Create new session with the new keyspace
	newSession, err := s.cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create session with keyspace %s: %w", keyspace, err)
	}

	// Update the session
	s.Session = newSession

	// Reinitialize schema cache for the new keyspace
	if s.schemaCache != nil {
		s.schemaCache = NewSchemaCache(s)
	}

	return nil
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home := os.Getenv("HOME")
		if home == "" {
			// Fallback for systems where HOME is not set
			if userHome, err := os.UserHomeDir(); err == nil {
				home = userHome
			}
		}
		if home != "" {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}

// createTLSConfig creates a TLS configuration based on the SSL settings
func createTLSConfig(sslConfig *config.SSLConfig, hostname string) (*tls.Config, error) {
	// Determine server name for hostname verification
	serverName := hostname
	if hostname != "" {
		// Strip port if present (hostname might be "host:port")
		if colonIdx := strings.LastIndex(hostname, ":"); colonIdx > 0 {
			serverName = hostname[:colonIdx]
		}
	}

	// When AllowLegacyCN is enabled, we need to bypass standard verification
	// and do manual verification in VerifyConnection
	skipVerify := sslConfig.InsecureSkipVerify || (sslConfig.AllowLegacyCN && sslConfig.HostVerification)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: skipVerify, // #nosec G402 - Configurable TLS verification
	}

	// Set ServerName for hostname verification (only when not using legacy CN verification)
	if sslConfig.HostVerification && !sslConfig.AllowLegacyCN && serverName != "" {
		tlsConfig.ServerName = serverName
	}

	// Load client certificate if provided
	if sslConfig.CertPath != "" && sslConfig.KeyPath != "" {
		certPath := expandPath(sslConfig.CertPath)
		keyPath := expandPath(sslConfig.KeyPath)
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if provided
	if sslConfig.CAPath != "" {
		caPath := expandPath(sslConfig.CAPath)
		caCert, err := os.ReadFile(caPath) // #nosec G304 - Path from trusted user configuration
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %v", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Manual verification for legacy CN certificates
	// We skip standard verification and manually check the certificate
	if sslConfig.AllowLegacyCN && sslConfig.HostVerification {
		tlsConfig.VerifyConnection = func(cs tls.ConnectionState) error {
			if len(cs.PeerCertificates) == 0 {
				return fmt.Errorf("no peer certificates")
			}

			// Build intermediates pool
			intermediates := x509.NewCertPool()
			for _, cert := range cs.PeerCertificates[1:] {
				intermediates.AddCert(cert)
			}

			// First try standard verification with SANs
			opts := x509.VerifyOptions{
				DNSName:       serverName,
				Intermediates: intermediates,
			}
			if tlsConfig.RootCAs != nil {
				opts.Roots = tlsConfig.RootCAs
			}

			_, err := cs.PeerCertificates[0].Verify(opts)
			if err == nil {
				return nil // Standard verification with SANs passed
			}

			// If standard verification failed, try legacy CN verification
			// First verify the certificate chain without hostname check
			optsNoHostname := x509.VerifyOptions{
				Intermediates: intermediates,
			}
			if tlsConfig.RootCAs != nil {
				optsNoHostname.Roots = tlsConfig.RootCAs
			}

			_, err = cs.PeerCertificates[0].Verify(optsNoHostname)
			if err != nil {
				return fmt.Errorf("certificate verification failed: %v", err)
			}

			// Chain is valid, now check CN
			cert := cs.PeerCertificates[0]
			if cert.Subject.CommonName == serverName {
				return nil // Legacy CN matches
			}

			return fmt.Errorf("certificate CN %q doesn't match expected hostname %q", cert.Subject.CommonName, serverName)
		}
	}

	return tlsConfig, nil
}
