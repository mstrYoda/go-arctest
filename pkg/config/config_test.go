package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yml")

	configContent := `
layers:
  - name: Domain
    pattern: "^domain/.*$"
  - name: Application
    pattern: "^application/.*$"
  - name: Infrastructure
    pattern: "^infrastructure/.*$"
  - name: Presentation
    pattern: "^presentation/.*$"

rules:
  - layer: Application
    dependsOn: Domain
  - layer: Infrastructure
    dependsOn: Domain
  - layer: Infrastructure
    dependsOn: Application
  - layer: Presentation
    dependsOn: Domain
  - layer: Presentation
    dependsOn: Application
  - layer: Presentation
    dependsOn: Infrastructure
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Load the config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the config
	if len(config.Layers) != 4 {
		t.Errorf("Expected 4 layers, got %d", len(config.Layers))
	}

	if len(config.Rules) != 6 {
		t.Errorf("Expected 6 rules, got %d", len(config.Rules))
	}

	// Check layer names
	expectedLayers := map[string]bool{
		"Domain":         true,
		"Application":    true,
		"Infrastructure": true,
		"Presentation":   true,
	}

	for _, layer := range config.Layers {
		if !expectedLayers[layer.Name] {
			t.Errorf("Unexpected layer name: %s", layer.Name)
		}
	}

	// Check rules
	for _, rule := range config.Rules {
		if rule.Layer == "Domain" {
			t.Errorf("Domain should not depend on any layer, but found rule: %s -> %s", rule.Layer, rule.DependsOn)
		}
	}
}

func TestValidateConfig(t *testing.T) {
	// Test with valid config
	validConfig := &Config{
		Layers: []LayerConfig{
			{Name: "Layer1", Pattern: "^layer1$"},
			{Name: "Layer2", Pattern: "^layer2$"},
		},
		Rules: []RuleConfig{
			{Layer: "Layer1", DependsOn: "Layer2"},
		},
	}

	if err := validateConfig(validConfig); err != nil {
		t.Errorf("Valid config should not return error, got: %v", err)
	}

	// Test with no layers
	noLayersConfig := &Config{
		Layers: []LayerConfig{},
		Rules:  []RuleConfig{},
	}

	if err := validateConfig(noLayersConfig); err == nil {
		t.Error("Expected error for config with no layers, got nil")
	}

	// Test with duplicate layer names
	duplicateLayersConfig := &Config{
		Layers: []LayerConfig{
			{Name: "Layer1", Pattern: "^layer1$"},
			{Name: "Layer1", Pattern: "^layer1_duplicate$"},
		},
		Rules: []RuleConfig{},
	}

	if err := validateConfig(duplicateLayersConfig); err == nil {
		t.Error("Expected error for config with duplicate layer names, got nil")
	}

	// Test with invalid regex pattern
	invalidPatternConfig := &Config{
		Layers: []LayerConfig{
			{Name: "Layer1", Pattern: "^layer1$"},
			{Name: "Layer2", Pattern: "[invalid"},
		},
		Rules: []RuleConfig{},
	}

	if err := validateConfig(invalidPatternConfig); err == nil {
		t.Error("Expected error for config with invalid regex pattern, got nil")
	}

	// Test with undefined layer in rule
	undefinedLayerConfig := &Config{
		Layers: []LayerConfig{
			{Name: "Layer1", Pattern: "^layer1$"},
		},
		Rules: []RuleConfig{
			{Layer: "Layer1", DependsOn: "UndefinedLayer"},
		},
	}

	if err := validateConfig(undefinedLayerConfig); err == nil {
		t.Error("Expected error for config with undefined layer in rule, got nil")
	}
}

func TestSaveConfig(t *testing.T) {
	// Create a config
	config := &Config{
		Layers: []LayerConfig{
			{Name: "Layer1", Pattern: "^layer1$"},
			{Name: "Layer2", Pattern: "^layer2$"},
		},
		Rules: []RuleConfig{
			{Layer: "Layer1", DependsOn: "Layer2"},
		},
	}

	// Save the config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "saved-config.yml")

	err := config.SaveConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load the saved config
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	// Verify the loaded config
	if len(loadedConfig.Layers) != len(config.Layers) {
		t.Errorf("Expected %d layers, got %d", len(config.Layers), len(loadedConfig.Layers))
	}

	if len(loadedConfig.Rules) != len(config.Rules) {
		t.Errorf("Expected %d rules, got %d", len(config.Rules), len(loadedConfig.Rules))
	}
}
