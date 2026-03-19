package logger

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateDefaults(t *testing.T) {
	opts := &LoggerOptions{}
	l := Create(opts)

	if l.Options.MessageColor != "white" {
		t.Errorf("Erwartete Farbe 'white', got '%s'", l.Options.MessageColor)
	}
	if l.Options.ErrorColor != "red" {
		t.Errorf("Erwartete Farbe 'red', got '%s'", l.Options.ErrorColor)
	}
	if l.Options.SuccessColor != "green" {
		t.Errorf("Erwartete Farbe 'green', got '%s'", l.Options.SuccessColor)
	}
	if l.Options.WarningColor != "yellow" {
		t.Errorf("Erwartete Farbe 'yellow', got '%s'", l.Options.WarningColor)
	}
}

func TestFormatting(t *testing.T) {

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	opts := &LoggerOptions{
		LogToFile: true,
	}
	l := Create(opts)

	testName := "Max Mustermann"
	l.LogErrorf("Fehler für %s", testName)

	output := buf.String()
	if !strings.Contains(output, "Fehler für Max Mustermann") {
		t.Errorf("Formatierung fehlgeschlagen. Got: %s", output)
	}
}

func TestFileLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	opts := &LoggerOptions{
		LogToFile: true,
		Filename:  logFile,
	}

	l := Create(opts)
	msg := "Test Nachricht für Datei"
	l.LogMessage(msg)

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatal("Log-Datei wurde nicht erstellt")
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), msg) {
		t.Errorf("Nachricht nicht in Datei gefunden. Dateiinhalt: %s", string(content))
	}
}

func TestColorSelection(t *testing.T) {
	opts := &LoggerOptions{
		LogToConsole: true,
		LogWithColor: true,
	}
	l := Create(opts)

	l.LogSuccess("Erfolg")
	l.LogWarning("Warnung")
	l.LogError("Fehler")
	l.LogMessage("Info")
}

func TestLogWithoutToFile(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	opts := &LoggerOptions{
		LogToFile: false,
	}
	l := Create(opts)
	l.LogMessage("Sollte nicht im Log-Buffer landen")

	if buf.Len() > 0 {
		t.Errorf("Es wurde etwas geloggt, obwohl LogToFile false war: %s", buf.String())
	}
}
