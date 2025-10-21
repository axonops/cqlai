package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	
	"github.com/axonops/cqlai/internal/logger"
)

// Config holds the application configuration
type Config struct {
	Host                string          `json:"host"`
	Port                int             `json:"port"`
	Keyspace            string          `json:"keyspace"`
	Username            string          `json:"username"`
	Password            string          `json:"password"`
	RequireConfirmation bool            `json:"requireConfirmation,omitempty"`
	Consistency         string          `json:"consistency,omitempty"`         // Default consistency level (e.g., "LOCAL_ONE", "QUORUM")
	PageSize            int             `json:"pageSize,omitempty"`
	MaxMemoryMB         int             `json:"maxMemoryMB,omitempty"`         // Max memory for results in MB (default: 10)
	ConnectTimeout      int             `json:"connectTimeout,omitempty"`      // Connection timeout in seconds
	RequestTimeout      int             `json:"requestTimeout,omitempty"`      // Request timeout in seconds
	Debug               bool            `json:"debug,omitempty"`               // Enable debug logging
	HistoryFile         string          `json:"historyFile,omitempty"`         // Path to CQL command history file
	AIHistoryFile       string          `json:"aiHistoryFile,omitempty"`       // Path to AI command history file
	SSL                 *SSLConfig      `json:"ssl,omitempty"`
	AI                  *AIConfig       `json:"ai,omitempty"`
	AuthProvider        *AuthProvider   `json:"authProvider,omitempty"`
}

// AuthProvider holds authentication provider configuration
type AuthProvider struct {
	Module    string `json:"module,omitempty"`    // e.g., "cassandra.auth"
	ClassName string `json:"className,omitempty"` // e.g., "PlainTextAuthProvider"
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
	Provider   string            `json:"provider"` // "mock", "openai", "anthropic", "gemini", "ollama", "openrouter"
	APIKey     string            `json:"apiKey"`   // General API key (overridden by provider-specific)
	Model      string            `json:"model"`    // General model (overridden by provider-specific)
	URL        string            `json:"url,omitempty"` // General URL (overridden by provider-specific)
	OpenAI     *AIProviderConfig `json:"openai,omitempty"`
	Anthropic  *AIProviderConfig `json:"anthropic,omitempty"`
	Gemini     *AIProviderConfig `json:"gemini,omitempty"`
	Ollama     *AIProviderConfig `json:"ollama,omitempty"`
	OpenRouter *AIProviderConfig `json:"openrouter,omitempty"`
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

// LoadConfig loads configuration from file and environment variables
// If customConfigPath is provided and not empty, it will be used instead of default locations
func LoadConfig(customConfigPath ...string) (*Config, error) {
	logger.DebugfToFile("Config", "Starting LoadConfig")

	config := &Config{
		Host:        "localhost",
		Port:        9042,
		MaxMemoryMB: 10, // Default to 10MB if not specified
	}

	logger.DebugfToFile("Config", "Default config initialized: host=%s, port=%d", config.Host, config.Port)

	// First, try to load CQLSHRC file
	cqlshrcPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".cassandra", "cqlshrc"),
		filepath.Join(os.Getenv("HOME"), ".cqlshrc"),
	}

	logger.DebugfToFile("Config", "Looking for cqlshrc files in: %v", cqlshrcPaths)

	for _, path := range cqlshrcPaths {
		logger.DebugfToFile("Config", "Attempting to load cqlshrc from: %s", path)
		if err := loadCQLSHRC(path, config); err == nil {
			logger.DebugfToFile("Config", "Successfully loaded cqlshrc from: %s", path)
			logger.DebugfToFile("Config", "Config after cqlshrc: host=%s, port=%d, username=%s, keyspace=%s",
				config.Host, config.Port, config.Username, config.Keyspace)
			break
		} else {
			logger.DebugfToFile("Config", "Failed to load cqlshrc from %s: %v", path, err)
		}
	}

	// Then check JSON config file locations (these will override CQLSHRC settings)
	var configPaths []string

	// If a custom config path is provided, use only that path
	if len(customConfigPath) > 0 && customConfigPath[0] != "" {
		configPaths = []string{customConfigPath[0]}
		logger.DebugfToFile("Config", "Using custom config path: %s", customConfigPath[0])
	} else {
		// Use default locations
		configPaths = []string{
			"cqlai.json",
			filepath.Join(os.Getenv("HOME"), ".cqlai.json"),
			filepath.Join(os.Getenv("HOME"), ".config", "cqlai", "config.json"),
		}
		logger.DebugfToFile("Config", "Looking for JSON config files in: %v", configPaths)
	}

	var configData []byte
	var err error
	var foundPath string

	for _, path := range configPaths {
		logger.DebugfToFile("Config", "Checking JSON config at: %s", path)
		configData, err = os.ReadFile(path) // #nosec G304 - Config file path is validated
		if err == nil {
			foundPath = path
			logger.DebugfToFile("Config", "Found JSON config at: %s", path)
			break
		} else {
			logger.DebugfToFile("Config", "No JSON config at %s: %v", path, err)
		}
	}

	// If custom config path was provided but not found, return an error
	if len(customConfigPath) > 0 && customConfigPath[0] != "" && foundPath == "" {
		return nil, fmt.Errorf("config file not found: %s", customConfigPath[0])
	}

	if foundPath != "" {
		if err := json.Unmarshal(configData, config); err != nil {
			logger.DebugfToFile("Config", "Failed to parse JSON config %s: %v", foundPath, err)
			return nil, fmt.Errorf("error parsing config file %s: %w", foundPath, err)
		}
		logger.DebugfToFile("Config", "Config after JSON: host=%s, port=%d, username=%s, keyspace=%s",
			config.Host, config.Port, config.Username, config.Keyspace)
	}

	// Override with environment variables
	logger.DebugfToFile("Config", "Applying environment variable overrides")
	OverrideWithEnvVars(config)

	logger.DebugfToFile("Config", "Final config: host=%s, port=%d, username=%s, keyspace=%s, hasPassword=%v",
		config.Host, config.Port, config.Username, config.Keyspace, config.Password != "")

	return config, nil
}

// OverrideWithEnvVars overrides configuration with environment variables
func OverrideWithEnvVars(config *Config) {
	// Connection settings
	if host := os.Getenv("CASSANDRA_HOST"); host != "" {
		config.Host = host
	}
	if host := os.Getenv("CQLAI_HOST"); host != "" {
		config.Host = host
	}

	if port := os.Getenv("CASSANDRA_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Port = p
		}
	}
	if port := os.Getenv("CQLAI_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Port = p
		}
	}

	if keyspace := os.Getenv("CASSANDRA_KEYSPACE"); keyspace != "" {
		config.Keyspace = keyspace
	}
	if keyspace := os.Getenv("CQLAI_KEYSPACE"); keyspace != "" {
		config.Keyspace = keyspace
	}

	if username := os.Getenv("CASSANDRA_USERNAME"); username != "" {
		config.Username = username
	}
	if username := os.Getenv("CQLAI_USERNAME"); username != "" {
		config.Username = username
	}

	if password := os.Getenv("CASSANDRA_PASSWORD"); password != "" {
		config.Password = password
	}
	if password := os.Getenv("CQLAI_PASSWORD"); password != "" {
		config.Password = password
	}

	// Page size setting
	if pageSize := os.Getenv("CQLAI_PAGE_SIZE"); pageSize != "" {
		if p, err := strconv.Atoi(pageSize); err == nil && p > 0 {
			config.PageSize = p
		}
	}

	// Max memory limit setting
	if maxMemory := os.Getenv("CQLAI_MAX_MEMORY_MB"); maxMemory != "" {
		if m, err := strconv.Atoi(maxMemory); err == nil && m > 0 {
			config.MaxMemoryMB = m
		}
	}

	// AI provider settings
	if provider := os.Getenv("AI_PROVIDER"); provider != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		config.AI.Provider = provider
	}
	if provider := os.Getenv("CQLAI_AI_PROVIDER"); provider != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		config.AI.Provider = provider
	}

	// OpenAI settings
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		if config.AI.OpenAI == nil {
			config.AI.OpenAI = &AIProviderConfig{}
		}
		config.AI.OpenAI.APIKey = apiKey
	}

	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		if config.AI.OpenAI == nil {
			config.AI.OpenAI = &AIProviderConfig{}
		}
		config.AI.OpenAI.Model = model
	}

	// Anthropic settings
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		if config.AI.Anthropic == nil {
			config.AI.Anthropic = &AIProviderConfig{}
		}
		config.AI.Anthropic.APIKey = apiKey
	}

	if model := os.Getenv("ANTHROPIC_MODEL"); model != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		if config.AI.Anthropic == nil {
			config.AI.Anthropic = &AIProviderConfig{}
		}
		config.AI.Anthropic.Model = model
	}

	// Gemini settings
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		if config.AI.Gemini == nil {
			config.AI.Gemini = &AIProviderConfig{}
		}
		config.AI.Gemini.APIKey = apiKey
	}

	// Ollama settings
	if url := os.Getenv("OLLAMA_URL"); url != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		if config.AI.Ollama == nil {
			config.AI.Ollama = &AIProviderConfig{}
		}
		config.AI.Ollama.URL = url
	}

	if model := os.Getenv("OLLAMA_MODEL"); model != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		if config.AI.Ollama == nil {
			config.AI.Ollama = &AIProviderConfig{}
		}
		config.AI.Ollama.Model = model
	}

	// OpenRouter settings
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		if config.AI.OpenRouter == nil {
			config.AI.OpenRouter = &AIProviderConfig{}
		}
		config.AI.OpenRouter.APIKey = apiKey
	}

	if model := os.Getenv("OPENROUTER_MODEL"); model != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		if config.AI.OpenRouter == nil {
			config.AI.OpenRouter = &AIProviderConfig{}
		}
		config.AI.OpenRouter.Model = model
	}

	if url := os.Getenv("OPENROUTER_URL"); url != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		if config.AI.OpenRouter == nil {
			config.AI.OpenRouter = &AIProviderConfig{}
		}
		config.AI.OpenRouter.URL = url
	}

	// General AI settings (fallback for any provider)
	if apiKey := os.Getenv("AI_API_KEY"); apiKey != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		config.AI.APIKey = apiKey
	}
	if apiKey := os.Getenv("CQLAI_AI_API_KEY"); apiKey != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		config.AI.APIKey = apiKey
	}

	if model := os.Getenv("AI_MODEL"); model != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		config.AI.Model = model
	}
	if model := os.Getenv("CQLAI_AI_MODEL"); model != "" {
		if config.AI == nil {
			config.AI = &AIConfig{}
		}
		config.AI.Model = model
	}
}

// ParseOutputFormat converts a string to OutputFormat
func ParseOutputFormat(format string) (OutputFormat, error) {
	switch strings.ToUpper(format) {
	case "TABLE":
		return OutputFormatTable, nil
	case "ASCII":
		return OutputFormatASCII, nil
	case "EXPAND":
		return OutputFormatExpand, nil
	case "JSON":
		return OutputFormatJSON, nil
	default:
		return "", fmt.Errorf("unknown output format: %s", format)
	}
}

// loadCQLSHRC loads configuration from a CQLSHRC file
func loadCQLSHRC(path string, config *Config) error {
	logger.DebugfToFile("CQLSHRC", "Attempting to open file: %s", path)
	
	file, err := os.Open(path) // #nosec G304 - Config file path is validated
	if err != nil {
		logger.DebugfToFile("CQLSHRC", "Failed to open file %s: %v", path, err)
		return err
	}
	defer file.Close()
	
	logger.DebugfToFile("CQLSHRC", "Successfully opened file: %s", path)

	scanner := bufio.NewScanner(file)
	currentSection := ""
	var credentialsPath string
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		// Log non-empty, non-comment lines
		if line != "" && !strings.HasPrefix(line, ";") && !strings.HasPrefix(line, "#") {
			logger.DebugfToFile("CQLSHRC", "Line %d: %s", lineNum, line)
		}

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.Trim(line, "[]"))
			logger.DebugfToFile("CQLSHRC", "Entering section: %s", currentSection)
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			logger.DebugfToFile("CQLSHRC", "Skipping invalid line %d (no '=' found): %s", lineNum, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		
		logger.DebugfToFile("CQLSHRC", "Section [%s], key=%s, value=%s", currentSection, key, value)

		// Map CQLSHRC values to config
		switch currentSection {
		case "connection":
			switch key {
			case "hostname":
				config.Host = value
				logger.DebugfToFile("CQLSHRC", "Set host to: %s", value)
			case "port":
				if port, err := strconv.Atoi(value); err == nil {
					config.Port = port
					logger.DebugfToFile("CQLSHRC", "Set port to: %d", port)
				} else {
					logger.DebugfToFile("CQLSHRC", "Failed to parse port value: %s", value)
				}
			case "ssl":
				if value == "true" || value == "1" {
					if config.SSL == nil {
						config.SSL = &SSLConfig{}
					}
					config.SSL.Enabled = true
					logger.DebugfToFile("CQLSHRC", "SSL enabled")
				}
			}
		case "authentication":
			switch key {
			case "credentials":
				credentialsPath = value
				logger.DebugfToFile("CQLSHRC", "Found credentials file path: %s", value)
			case "keyspace":
				config.Keyspace = value
				logger.DebugfToFile("CQLSHRC", "Set keyspace to: %s", value)
			case "username":
				config.Username = value
				logger.DebugfToFile("CQLSHRC", "Set username to: %s", value)
			case "password":
				config.Password = value
				logger.DebugfToFile("CQLSHRC", "Set password (hidden)")
			}
		case "auth_provider":
			if config.AuthProvider == nil {
				config.AuthProvider = &AuthProvider{}
			}
			switch key {
			case "module":
				config.AuthProvider.Module = value
				logger.DebugfToFile("CQLSHRC", "Set auth module to: %s", value)
			case "classname":
				config.AuthProvider.ClassName = value
				logger.DebugfToFile("CQLSHRC", "Set auth classname to: %s", value)
			case "username":
				config.Username = value
				logger.DebugfToFile("CQLSHRC", "Set username to: %s", value)
			case "password":
				config.Password = value
				logger.DebugfToFile("CQLSHRC", "Set password (hidden)")
			}
		case "ssl":
			if config.SSL == nil {
				config.SSL = &SSLConfig{}
			}
			// Any key in [ssl] section means SSL should be enabled
			config.SSL.Enabled = true
			switch key {
			case "factory":
				// Ignore factory setting - we handle SSL ourselves
				logger.DebugfToFile("CQLSHRC", "SSL enabled (factory specified)")
			case "certfile":
				// Expand ~ to home directory
				if strings.HasPrefix(value, "~") {
					value = filepath.Join(os.Getenv("HOME"), value[1:])
				}
				config.SSL.CAPath = value
				logger.DebugfToFile("CQLSHRC", "Set CA path to: %s", value)
			case "userkey":
				if strings.HasPrefix(value, "~") {
					value = filepath.Join(os.Getenv("HOME"), value[1:])
				}
				config.SSL.KeyPath = value
				logger.DebugfToFile("CQLSHRC", "Set key path to: %s", value)
			case "usercert":
				if strings.HasPrefix(value, "~") {
					value = filepath.Join(os.Getenv("HOME"), value[1:])
				}
				config.SSL.CertPath = value
				logger.DebugfToFile("CQLSHRC", "Set cert path to: %s", value)
			case "validate":
				if value == "false" || value == "0" {
					config.SSL.InsecureSkipVerify = true
					config.SSL.HostVerification = false
					logger.DebugfToFile("CQLSHRC", "Set InsecureSkipVerify to true and HostVerification to false")
				} else {
					config.SSL.HostVerification = true
					logger.DebugfToFile("CQLSHRC", "Set HostVerification to true")
				}
			}
		}
	}

	// If a credentials file was specified, try to load it
	if credentialsPath != "" {
		logger.DebugfToFile("CQLSHRC", "Loading credentials file: %s", credentialsPath)
		if err := loadCredentialsFile(credentialsPath, config); err != nil {
			logger.DebugfToFile("CQLSHRC", "Failed to load credentials file: %v", err)
		} else {
			logger.DebugfToFile("CQLSHRC", "Successfully loaded credentials file")
		}
	}
	
	if err := scanner.Err(); err != nil {
		logger.DebugfToFile("CQLSHRC", "Scanner error: %v", err)
		return err
	}
	
	logger.DebugfToFile("CQLSHRC", "Finished loading cqlshrc: host=%s, port=%d, username=%s, hasPassword=%v", 
		config.Host, config.Port, config.Username, config.Password != "")

	return nil
}

// loadCredentialsFile loads username/password from a credentials file
// The format is typically:
// [auth_provider_classname]
// username = user
// password = pass
func loadCredentialsFile(path string, config *Config) error {
	logger.DebugfToFile("Credentials", "Loading credentials file: %s", path)
	
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		path = filepath.Join(os.Getenv("HOME"), path[1:])
		logger.DebugfToFile("Credentials", "Expanded path to: %s", path)
	}

	file, err := os.Open(path) // #nosec G304 - Config file path is validated
	if err != nil {
		logger.DebugfToFile("Credentials", "Failed to open credentials file %s: %v", path, err)
		return err
	}
	defer file.Close()
	
	logger.DebugfToFile("Credentials", "Successfully opened credentials file")

	scanner := bufio.NewScanner(file)
	inAuthSection := false
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		
		logger.DebugfToFile("Credentials", "Line %d: %s", lineNum, line)

		// Check for section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.ToLower(strings.Trim(line, "[]"))
			// Look for PlainTextAuthProvider or similar auth sections
			inAuthSection = strings.Contains(section, "auth")
			logger.DebugfToFile("Credentials", "Section [%s], inAuthSection=%v", section, inAuthSection)
			continue
		}

		if !inAuthSection {
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			logger.DebugfToFile("Credentials", "Skipping invalid line %d (no '=' found): %s", lineNum, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}

		switch key {
		case "username":
			config.Username = value
			logger.DebugfToFile("Credentials", "Set username to: %s", value)
		case "password":
			config.Password = value
			logger.DebugfToFile("Credentials", "Set password (hidden)")
		default:
			logger.DebugfToFile("Credentials", "Ignoring unknown key: %s", key)
		}
	}
	
	if err := scanner.Err(); err != nil {
		logger.DebugfToFile("Credentials", "Scanner error: %v", err)
		return err
	}
	
	logger.DebugfToFile("Credentials", "Finished loading credentials: username=%s, hasPassword=%v", 
		config.Username, config.Password != "")

	return nil
}
