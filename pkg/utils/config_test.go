package utils

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Сохраняем оригинальные env переменные
	originalToken := os.Getenv("TELEGRAM_TOKEN")
	originalDBURL := os.Getenv("DATABASE_URL")
	originalDebug := os.Getenv("DEBUG")
	originalAdminIDs := os.Getenv("ADMIN_IDS")

	// Восстанавливаем env после теста
	defer func() {
		os.Setenv("TELEGRAM_TOKEN", originalToken)
		os.Setenv("DATABASE_URL", originalDBURL)
		os.Setenv("DEBUG", originalDebug)
		os.Setenv("ADMIN_IDS", originalAdminIDs)
	}()

	tests := []struct {
		name           string
		setupEnv       func()
		wantErr        bool
		expectToken    string
		expectDBURL    string
		expectDebug    bool
		expectAdminIDs []int64
	}{
		{
			name: "Valid config with all parameters",
			setupEnv: func() {
				os.Setenv("TELEGRAM_TOKEN", "test_token_123")
				os.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
				os.Setenv("DEBUG", "true")
				os.Setenv("ADMIN_IDS", "123,456,789")
			},
			wantErr:        false,
			expectToken:    "test_token_123",
			expectDBURL:    "postgres://user:pass@localhost/db",
			expectDebug:    true,
			expectAdminIDs: []int64{123, 456, 789},
		},
		{
			name: "Valid config without optional parameters",
			setupEnv: func() {
				os.Setenv("TELEGRAM_TOKEN", "simple_token")
				os.Setenv("DATABASE_URL", "sqlite://local.db")
				os.Unsetenv("DEBUG")
				os.Unsetenv("ADMIN_IDS")
			},
			wantErr:        false,
			expectToken:    "simple_token",
			expectDBURL:    "sqlite://local.db",
			expectDebug:    false,
			expectAdminIDs: []int64{},
		},
		{
			name: "Missing Telegram token",
			setupEnv: func() {
				os.Unsetenv("TELEGRAM_TOKEN")
				os.Setenv("DATABASE_URL", "valid_url")
			},
			wantErr:     true,
			expectToken: "", // не важно, так как ожидаем ошибку
			expectDBURL: "valid_url",
		},
		{
			name: "Missing database URL",
			setupEnv: func() {
				os.Setenv("TELEGRAM_TOKEN", "valid_token")
				os.Unsetenv("DATABASE_URL")
			},
			wantErr:     true,
			expectToken: "valid_token",
			expectDBURL: "", // не важно, так как ожидаем ошибку
		},
		{
			name: "Debug false by default",
			setupEnv: func() {
				os.Setenv("TELEGRAM_TOKEN", "token")
				os.Setenv("DATABASE_URL", "url")
				os.Setenv("DEBUG", "false")
			},
			wantErr:        false,
			expectToken:    "token",
			expectDBURL:    "url",
			expectDebug:    false,
			expectAdminIDs: []int64{},
		},
		{
			name: "Debug true with various truthy values",
			setupEnv: func() {
				os.Setenv("TELEGRAM_TOKEN", "token")
				os.Setenv("DATABASE_URL", "url")
				os.Setenv("DEBUG", "1")
			},
			wantErr:        false,
			expectToken:    "token",
			expectDBURL:    "url",
			expectDebug:    true,
			expectAdminIDs: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment before test
			os.Unsetenv("TELEGRAM_TOKEN")
			os.Unsetenv("DATABASE_URL")
			os.Unsetenv("DEBUG")
			os.Unsetenv("ADMIN_IDS")

			// Setup test environment
			tt.setupEnv()

			config, err := LoadConfig()

			if tt.wantErr {
				if err == nil {
					t.Errorf("LoadConfig() expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("LoadConfig() unexpected error: %v", err)
				return
			}

			if config.TelegramToken != tt.expectToken {
				t.Errorf("TelegramToken = %v, want %v", config.TelegramToken, tt.expectToken)
			}

			if config.DatabaseURL != tt.expectDBURL {
				t.Errorf("DatabaseURL = %v, want %v", config.DatabaseURL, tt.expectDBURL)
			}

			if config.Debug != tt.expectDebug {
				t.Errorf("Debug = %v, want %v", config.Debug, tt.expectDebug)
			}

			// Check AdminIDs length
			if len(config.AdminIDs) != len(tt.expectAdminIDs) {
				t.Errorf("AdminIDs length = %v, want %v", len(config.AdminIDs), len(tt.expectAdminIDs))
			}

			// Check individual AdminIDs
			for i, id := range config.AdminIDs {
				if i >= len(tt.expectAdminIDs) {
					break
				}
				if id != tt.expectAdminIDs[i] {
					t.Errorf("AdminIDs[%d] = %v, want %v", i, id, tt.expectAdminIDs[i])
				}
			}
		})
	}
}

func TestParseAdminIDs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int64
	}{
		{
			name:     "Single ID",
			input:    "123",
			expected: []int64{123},
		},
		{
			name:     "Multiple IDs",
			input:    "123,456,789",
			expected: []int64{123, 456, 789},
		},
		{
			name:     "IDs with spaces",
			input:    "123, 456, 789",
			expected: []int64{123, 456, 789},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []int64{},
		},
		{
			name:     "String with empty elements",
			input:    "123,,456,",
			expected: []int64{123, 456},
		},
		{
			name:     "Invalid ID skipped",
			input:    "123,abc,456",
			expected: []int64{123, 456},
		},
		{
			name:     "Whitespace only",
			input:    "   , , ",
			expected: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAdminIDs(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("parseAdminIDs(%q) length = %v, want %v", tt.input, len(result), len(tt.expected))
				return
			}

			for i, id := range result {
				if id != tt.expected[i] {
					t.Errorf("parseAdminIDs(%q)[%d] = %v, want %v", tt.input, i, id, tt.expected[i])
				}
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	originalValue := os.Getenv("TEST_VAR")
	defer os.Setenv("TEST_VAR", originalValue)

	tests := []struct {
		name       string
		envValue   string
		defaultVal string
		expected   string
	}{
		{
			name:       "Environment variable set",
			envValue:   "actual_value",
			defaultVal: "default_value",
			expected:   "actual_value",
		},
		{
			name:       "Environment variable empty",
			envValue:   "",
			defaultVal: "default_value",
			expected:   "default_value",
		},
		{
			name:       "Environment variable not set",
			envValue:   "DELETE", // Special value to unset
			defaultVal: "default_value",
			expected:   "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue == "DELETE" {
				os.Unsetenv("TEST_VAR")
			} else {
				os.Setenv("TEST_VAR", tt.envValue)
			}

			result := getEnv("TEST_VAR", tt.defaultVal)
			if result != tt.expected {
				t.Errorf("getEnv() = %v, want %v", result, tt.expected)
			}
		})
	}
}
