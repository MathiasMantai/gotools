package env

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {

	testCases := []struct {
		name         string
		key          string
		defaultValue string
		expected     string
		setenv       bool
	}{
		{"Value exists", "TEST_VAR1", "", "Hallo Welt", true},
		{"Value does not exist", "TEST_VAR2", "default_value", "default_value", false},
		{"Empty value", "TEST_VAR_EMPTY", "", "", true},
		{"Value does not exist and default is empty", "TEST_VAR_EMPTY_DEFAULT", "", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setenv {
				os.Setenv(tc.key, tc.expected)
				defer os.Unsetenv(tc.key)
			}
			if got := Get(tc.key, tc.defaultValue); got != tc.expected {
				t.Errorf("=> Get: expected: %v, actual: %v\n", got, tc.expected)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	envFilePath := filepath.Join(tmpDir, ".env")

	testRows := []string{
		"# comment",
		"SERVER=192.168.44.30",
		"USER=terminal_user",
	}

	err := os.WriteFile(envFilePath, []byte(strings.Join(testRows, "\n")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Successfully loaded env file", func(t *testing.T) {
		//make sure none of the values are loaded into env
		for _, testRow := range testRows {
			if strings.HasPrefix(testRow, "#") {
				continue
			}

			tmpParts := strings.Split(testRow, "=")

			if len(tmpParts) != 2 {
				t.Fatal(errors.New("invalid testrow"))
			}

			os.Unsetenv(strings.TrimSpace(tmpParts[0]))
		}

		err := Load([]string{
			envFilePath,
		})
		if err != nil {
			t.Errorf("Load(): error while loading env file: %v", err)
		}

		if os.Getenv("SERVER") != "192.168.44.30" {
			t.Errorf("=> Load(): expected: %v. actual: %v", "192.168.44.30", os.Getenv("SERVER"))
		}

		if os.Getenv("USER") != "terminal_user" {
			t.Errorf("=> Load(): expected: %v. actual: %v", "terminal_user", os.Getenv("USER"))
		}

	})
}
