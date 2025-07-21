package main

import (
	"os"
	"testing"
)

func TestReadConfig_ValidFile(t *testing.T) {
	yamlData := `
Rabbit:
  Channel: "my-channel"
  Host: "localhost"
  Username: "guest"
  Password: "guest"
  Port: 5672
Web:
  Host: "0.0.0.0"
  Port: 8080
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlData)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := ReadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("ReadConfig returned error: %v", err)
	}

	if cfg.Rabbit.Channel != "my-channel" {
		t.Errorf("Expected Rabbit.Channel 'my-channel', got '%s'", cfg.Rabbit.Channel)
	}
	if cfg.Web.Port != 8080 {
		t.Errorf("Expected Web.Port 8080, got %d", cfg.Web.Port)
	}
	if cfg.Web.Host != "0.0.0.0" {
		t.Errorf("Expected Web.Host '0.0.0.0', got '%s'", cfg.Web.Host)
	}
}

func TestReadConfig_FileNotFound(t *testing.T) {
	_, err := ReadConfig("nonexistent.yaml")
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}
}

func TestReadConfig_InvalidYAML(t *testing.T) {
	invalidYAML := `
Rabbit:
  Channel: "my-channel"
  Host: localhost
  Username: guest
  Password: guest
  Port: not-a-number
Web:
  Host: 0.0.0.0
  Port: 8080
`

	tmpFile, err := os.CreateTemp("", "invalid-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(invalidYAML)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	_, err = ReadConfig(tmpFile.Name())
	if err == nil {
		t.Fatal("Expected YAML unmarshal error, got nil")
	}
}

