// Copyright (c) 2025 Thorium

package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"thorium-cli/internal/config"
)

// Execute runs a SQL statement against the database
func Execute(db config.DBConfig, sqlContent string) error {
	conn, err := connect(db)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Split on semicolons to execute multiple statements
	// This is a simple approach - may need improvement for complex SQL
	statements := splitStatements(sqlContent)

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Remove leading comment lines from the statement
		stmt = stripLeadingComments(stmt)
		if stmt == "" {
			continue
		}

		_, err := conn.Exec(stmt)
		if err != nil {
			return fmt.Errorf("execute SQL: %w\nStatement: %s", err, truncate(stmt, 200))
		}
	}

	return nil
}

// Query runs a SQL query and returns the result
func Query(db config.DBConfig, sqlQuery string) (*sql.Rows, error) {
	conn, err := connect(db)
	if err != nil {
		return nil, err
	}
	// Note: caller must close rows and connection

	return conn.Query(sqlQuery)
}

// QueryValue runs a SQL query and returns a single value
func QueryValue(db config.DBConfig, sqlQuery string) (string, error) {
	conn, err := connect(db)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	var value string
	err = conn.QueryRow(sqlQuery).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return value, nil
}

// connect creates a database connection
func connect(db config.DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true",
		db.User, db.Password, db.Host, db.Port, db.Name)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return conn, nil
}

// splitStatements splits SQL content into individual statements
func splitStatements(content string) []string {
	// Remove comments and split on semicolons
	var statements []string
	var current strings.Builder
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(content); i++ {
		c := content[i]

		// Handle string literals
		if (c == '\'' || c == '"') && (i == 0 || content[i-1] != '\\') {
			if !inString {
				inString = true
				stringChar = c
			} else if c == stringChar {
				inString = false
			}
		}

		// Handle statement terminator
		if c == ';' && !inString {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
			continue
		}

		current.WriteByte(c)
	}

	// Add final statement if any
	stmt := strings.TrimSpace(current.String())
	if stmt != "" {
		statements = append(statements, stmt)
	}

	return statements
}

// stripLeadingComments removes leading comment lines from a SQL statement
func stripLeadingComments(stmt string) string {
	lines := strings.Split(stmt, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and comment-only lines at the start
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			if len(result) == 0 {
				continue // Skip leading comments
			}
		}
		result = append(result, line)
	}
	return strings.TrimSpace(strings.Join(result, "\n"))
}

// truncate truncates a string to maxLen
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// CreateDatabase creates a database if it doesn't exist
func CreateDatabase(dbConfig config.DBConfig) error {
	// Connect to MySQL without a database to manage databases
	connConfig := dbConfig
	connConfig.Name = ""
	
	conn, err := connect(connConfig)
	if err != nil {
		return fmt.Errorf("connect to MySQL: %w", err)
	}
	defer conn.Close()

	// Create database if it doesn't exist
	_, err = conn.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbConfig.Name))
	if err != nil {
		return fmt.Errorf("create database %s: %w", dbConfig.Name, err)
	}

	return nil
}

// InitializeThoriumDatabases creates all databases required by Thorium
func InitializeThoriumDatabases(cfg config.Config) error {
	databases := []struct {
		name   string
		config config.DBConfig
	}{
		{"dbc", cfg.Databases.DBC},
		{"dbc_source", cfg.Databases.DBCSource},
		{"world", cfg.Databases.World},
	}

	for _, db := range databases {
		if err := CreateDatabase(db.config); err != nil {
			return fmt.Errorf("create %s database: %w", db.name, err)
		}
	}

	return nil
}
