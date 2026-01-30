// Copyright (c) 2025 Thorium

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Config is the main thorium configuration
type Config struct {
	// Computed paths
	WorkspaceRoot string // Directory containing config.json
	ConfigPath    string // Path to this config file

	// WoW Client
	WoTLK WoTLKConfig `json:"wotlk"`

	// Databases
	Databases DatabasesConfig `json:"databases"`

	// Server paths
	Server ServerConfig `json:"server"`

	// Output settings
	Output OutputConfig `json:"output"`

}

// WoTLKConfig holds WoW client settings
type WoTLKConfig struct {
	Path   string `json:"path"`
	Locale string `json:"locale"`
}

// DatabasesConfig holds database connection settings
type DatabasesConfig struct {
	DBC   DBConfig `json:"dbc"`
	World DBConfig `json:"world"`
}

// DBConfig holds connection info for a single database
type DBConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Name     string `json:"name"`
}

// ServerConfig holds server paths
type ServerConfig struct {
	DBCPath string `json:"dbc_path"` // Where to copy DBCs for server
}

// OutputConfig holds output file settings
type OutputConfig struct {
	DBCMPQ    string `json:"dbc_mpq"`
	LuaXMLMPQ string `json:"luaxml_mpq"`
}


// Load loads and parses the config file, expanding environment variables
func Load(path string) (*Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve config path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	// Expand environment variables in the JSON
	expanded := expandEnvVars(string(data))

	var cfg Config
	if err := json.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Set computed paths
	cfg.ConfigPath = absPath
	cfg.WorkspaceRoot = filepath.Dir(absPath)

	// Apply defaults
	cfg.applyDefaults()

	return &cfg, nil
}

// expandEnvVars expands ${VAR} and ${VAR:-default} patterns
func expandEnvVars(s string) string {
	// Pattern: ${VAR:-default} or ${VAR}
	re := regexp.MustCompile(`\$\{([^}:]+)(?::-([^}]*))?\}`)

	return re.ReplaceAllStringFunc(s, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		varName := parts[1]
		defaultVal := ""
		if len(parts) >= 3 {
			defaultVal = parts[2]
		}

		if val := os.Getenv(varName); val != "" {
			return val
		}
		return defaultVal
	})
}

// applyDefaults sets default values for unset fields
func (c *Config) applyDefaults() {
	// WoTLK defaults
	if c.WoTLK.Locale == "" {
		c.WoTLK.Locale = "enUS"
	}

	// Database defaults
	if c.Databases.DBC.Host == "" {
		c.Databases.DBC.Host = getEnvOrDefault("MYSQL_HOST", "127.0.0.1")
	}
	if c.Databases.DBC.Port == "" {
		c.Databases.DBC.Port = getEnvOrDefault("MYSQL_PORT", "3306")
	}
	if c.Databases.DBC.Name == "" {
		c.Databases.DBC.Name = "dbc"
	}
	if c.Databases.DBC.User == "" {
		c.Databases.DBC.User = "root"
	}

	if c.Databases.World.Host == "" {
		c.Databases.World.Host = getEnvOrDefault("MYSQL_HOST", "127.0.0.1")
	}
	if c.Databases.World.Port == "" {
		c.Databases.World.Port = getEnvOrDefault("MYSQL_PORT", "3306")
	}
	if c.Databases.World.Name == "" {
		c.Databases.World.Name = "world"
	}
	if c.Databases.World.User == "" {
		c.Databases.World.User = "trinity"
	}
	if c.Databases.World.Password == "" {
		c.Databases.World.Password = "trinity"
	}

	// Output defaults
	if c.Output.DBCMPQ == "" {
		c.Output.DBCMPQ = "patch-T.MPQ"
	}
	if c.Output.LuaXMLMPQ == "" {
		c.Output.LuaXMLMPQ = "patch-{locale}-T.MPQ"
	}

}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// Path helpers - all relative to workspace root

// GetSharedPath returns the path to shared directory
func (c *Config) GetSharedPath() string {
	return filepath.Join(c.WorkspaceRoot, "shared")
}

// GetDBCSourcePath returns the path to source DBC files
func (c *Config) GetDBCSourcePath() string {
	return filepath.Join(c.WorkspaceRoot, "shared", "dbc", "dbc_source")
}

// GetDBCOutPath returns the path to exported DBC files
func (c *Config) GetDBCOutPath() string {
	return filepath.Join(c.WorkspaceRoot, "shared", "dbc", "dbc_out")
}

// GetDBCMetaPath returns empty string - meta files are now embedded in the binary
func (c *Config) GetDBCMetaPath() string {
	return "" // Meta files are embedded, not on disk
}

// GetLuaXMLSourcePath returns the path to source LuaXML files
func (c *Config) GetLuaXMLSourcePath() string {
	return filepath.Join(c.WorkspaceRoot, "shared", "luaxml", "luaxml_source")
}

// GetLuaXMLOutPath returns the path to modified LuaXML files
func (c *Config) GetLuaXMLOutPath() string {
	return filepath.Join(c.WorkspaceRoot, "shared", "luaxml", "luaxml_out")
}

// GetModsPath returns the path to mods directory
func (c *Config) GetModsPath() string {
	return filepath.Join(c.WorkspaceRoot, "mods")
}

// GetAppliedMigrationsPath returns the path to track applied migrations
func (c *Config) GetAppliedMigrationsPath() string {
	return filepath.Join(c.WorkspaceRoot, "shared", "migrations_applied")
}

// GetMPQName returns the MPQ name with locale substituted
func (c *Config) GetMPQName(template string) string {
	return strings.ReplaceAll(template, "{locale}", c.WoTLK.Locale)
}

// GetClientDataPath returns the path to WoW client Data directory
func (c *Config) GetClientDataPath() string {
	return filepath.Join(c.WoTLK.Path, "Data")
}

// GetClientLocalePath returns the path to WoW client locale directory
func (c *Config) GetClientLocalePath() string {
	return filepath.Join(c.WoTLK.Path, "Data", c.WoTLK.Locale)
}

// DefaultConfig returns a config with sensible defaults for scaffolding
func DefaultConfig() *Config {
	return &Config{
		WoTLK: WoTLKConfig{
			Path:   "${WOTLK_PATH}",
			Locale: "enUS",
		},
		Databases: DatabasesConfig{
			DBC: DBConfig{
				User:     "root",
				Password: "",
				Host:     "${MYSQL_HOST:-127.0.0.1}",
				Port:     "${MYSQL_PORT:-3306}",
				Name:     "dbc",
			},
			World: DBConfig{
				User:     "trinity",
				Password: "trinity",
				Host:     "${MYSQL_HOST:-127.0.0.1}",
				Port:     "${MYSQL_PORT:-3306}",
				Name:     "world",
			},
		},
		Server: ServerConfig{
			DBCPath: "${TC_SERVER_PATH}/data/dbc",
		},
		Output: OutputConfig{
			DBCMPQ:    "patch-T.MPQ",
			LuaXMLMPQ: "patch-{locale}-T.MPQ",
		},
	}
}

// WriteConfig writes a config to a file
func WriteConfig(cfg *Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
