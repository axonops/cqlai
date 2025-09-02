package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds the application configuration
type Config struct {
	Host                string          `json:"host"`
	Port                int             `json:"port"`
	Keyspace            string          `json:"keyspace"`
	Username            string          `json:"username"`
	Password            string          `json:"password"`
	RequireConfirmation bool            `json:"requireConfirmation,omitempty"`
	PageSize            int             `json:"pageSize,omitempty"`
	ConnectTimeout      int             `json:"connectTimeout,omitempty"`      // Connection timeout in seconds
	RequestTimeout      int             `json:"requestTimeout,omitempty"`      // Request timeout in seconds
	Debug               bool            `json:"debug,omitempty"`               // Enable debug logging
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

// LoadConfig loads configuration from file and environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		Host: "localhost",
		Port: 9042,
	}

	// First, try to load CQLSHRC file
	cqlshrcPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".cassandra", "cqlshrc"),
		filepath.Join(os.Getenv("HOME"), ".cqlshrc"),
	}

	for _, path := range cqlshrcPaths {
		if err := loadCQLSHRC(path, config); err == nil {
			break
		}
	}

	// Then check JSON config file locations (these will override CQLSHRC settings)
	configPaths := []string{
		"cqlai.json",
		filepath.Join(os.Getenv("HOME"), ".cqlai.json"),
		filepath.Join(os.Getenv("HOME"), ".config", "cqlai", "config.json"),
		"/etc/cqlai/config.json",
	}

	var configData []byte
	var err error
	var foundPath string

	for _, path := range configPaths {
		configData, err = os.ReadFile(path) // #nosec G304 - Config file path is validated
		if err == nil {
			foundPath = path
			break
		}
	}

	if foundPath != "" {
		if err := json.Unmarshal(configData, config); err != nil {
			return nil, fmt.Errorf("error parsing config file %s: %w", foundPath, err)
		}
	}

	// Override with environment variables
	OverrideWithEnvVars(config)

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
	file, err := os.Open(path) // #nosec G304 - Config file path is validated
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentSection := ""
	var credentialsPath string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.Trim(line, "[]"))
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}

		// Map CQLSHRC values to config
		switch currentSection {
		case "connection":
			switch key {
			case "hostname":
				config.Host = value
			case "port":
				if port, err := strconv.Atoi(value); err == nil {
					config.Port = port
				}
			case "ssl":
				if value == "true" || value == "1" {
					if config.SSL == nil {
						config.SSL = &SSLConfig{}
					}
					config.SSL.Enabled = true
				}
			}
		case "authentication":
			switch key {
			case "credentials":
				credentialsPath = value
			case "keyspace":
				config.Keyspace = value
			}
		case "auth_provider":
			if config.AuthProvider == nil {
				config.AuthProvider = &AuthProvider{}
			}
			switch key {
			case "module":
				config.AuthProvider.Module = value
			case "classname":
				config.AuthProvider.ClassName = value
			case "username":
				config.Username = value
				// Note: password is typically not stored in cqlshrc for security reasons
				// It's usually in a separate credentials file or provided via prompt
			}
		case "ssl":
			if config.SSL == nil {
				config.SSL = &SSLConfig{}
			}
			switch key {
			case "certfile":
				// Expand ~ to home directory
				if strings.HasPrefix(value, "~") {
					value = filepath.Join(os.Getenv("HOME"), value[1:])
				}
				config.SSL.CAPath = value
			case "userkey":
				if strings.HasPrefix(value, "~") {
					value = filepath.Join(os.Getenv("HOME"), value[1:])
				}
				config.SSL.KeyPath = value
			case "usercert":
				if strings.HasPrefix(value, "~") {
					value = filepath.Join(os.Getenv("HOME"), value[1:])
				}
				config.SSL.CertPath = value
			case "validate":
				if value == "false" || value == "0" {
					config.SSL.InsecureSkipVerify = true
				}
			}
		}
	}

	// If a credentials file was specified, try to load it
	if credentialsPath != "" {
		_ = loadCredentialsFile(credentialsPath, config)
	}

	return scanner.Err()
}

// loadCredentialsFile loads username/password from a credentials file
// The format is typically:
// [auth_provider_classname]
// username = user
// password = pass
func loadCredentialsFile(path string, config *Config) error {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		path = filepath.Join(os.Getenv("HOME"), path[1:])
	}

	file, err := os.Open(path) // #nosec G304 - Config file path is validated
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inAuthSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.ToLower(strings.Trim(line, "[]"))
			// Look for PlainTextAuthProvider or similar auth sections
			inAuthSection = strings.Contains(section, "auth")
			continue
		}

		if !inAuthSection {
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
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
		case "password":
			config.Password = value
		}
	}

	return scanner.Err()
}
