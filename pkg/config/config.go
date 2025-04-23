package config

import (
	"fmt"
	"os"
	"regexp"

	"github.com/mstrYoda/go-arctest/pkg/arctest"
	"gopkg.in/yaml.v3"
)

// Config represents the YAML configuration for architecture tests
type Config struct {
	Layers                     []LayerConfig                     `yaml:"layers"`
	Rules                      []RuleConfig                      `yaml:"rules"`
	InterfaceRules             []InterfaceRuleConfig             `yaml:"interfaceRules,omitempty"`
	ParameterRules             []ParameterRuleConfig             `yaml:"parameterRules,omitempty"`
	LayerSpecificRules         []LayerSpecificRuleConfig         `yaml:"layerSpecificRules,omitempty"`
	DirectLayerDependencyRules []DirectLayerDependencyRuleConfig `yaml:"directLayerDependencyRules,omitempty"`
}

// LayerConfig represents a layer in the architecture
type LayerConfig struct {
	Name    string `yaml:"name"`
	Pattern string `yaml:"pattern"`
}

// RuleConfig represents a dependency rule between layers
type RuleConfig struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// InterfaceRuleConfig represents a rule for interface implementations
type InterfaceRuleConfig struct {
	StructPattern    string `yaml:"structPattern"`
	InterfacePattern string `yaml:"interfacePattern"`
}

// ParameterRuleConfig represents a rule for method parameters
type ParameterRuleConfig struct {
	StructPattern        string `yaml:"structPattern"`
	MethodPattern        string `yaml:"methodPattern"`
	ParameterTypePattern string `yaml:"parameterTypePattern"`
	ShouldUseInterface   bool   `yaml:"shouldUseInterface"`
}

// LayerSpecificRuleConfig represents a rule specific to a layer
type LayerSpecificRuleConfig struct {
	Layer      string            `yaml:"layer"`
	RuleType   string            `yaml:"ruleType"` // "dependency", "interface", or "parameter"
	Parameters map[string]string `yaml:"parameters"`
}

// DirectLayerDependencyRuleConfig represents a direct dependency rule between layers
type DirectLayerDependencyRuleConfig struct {
	SourceLayer string `yaml:"sourceLayer"`
	TargetLayer string `yaml:"targetLayer"`
	Allowed     bool   `yaml:"allowed"`
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate the configuration
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	if len(config.Layers) == 0 {
		return fmt.Errorf("no layers defined in configuration")
	}

	// Check for duplicate layer names
	layerNames := make(map[string]bool)
	for _, layer := range config.Layers {
		if layer.Name == "" {
			return fmt.Errorf("layer name cannot be empty")
		}
		if layer.Pattern == "" {
			return fmt.Errorf("layer pattern cannot be empty for layer %s", layer.Name)
		}
		if layerNames[layer.Name] {
			return fmt.Errorf("duplicate layer name: %s", layer.Name)
		}
		layerNames[layer.Name] = true

		// Validate the pattern is a valid regex
		if _, err := regexp.Compile(layer.Pattern); err != nil {
			return fmt.Errorf("invalid pattern for layer %s: %w", layer.Name, err)
		}
	}

	// Validate rules
	for _, rule := range config.Rules {
		if rule.From == "" {
			return fmt.Errorf("rule 'from' cannot be empty")
		}
		if rule.To == "" {
			return fmt.Errorf("rule 'to' cannot be empty")
		}
		if !layerNames[rule.From] {
			return fmt.Errorf("rule references undefined layer: %s", rule.From)
		}
		if !layerNames[rule.To] {
			return fmt.Errorf("rule references undefined layer: %s", rule.To)
		}
	}

	// Validate interface rules
	for i, rule := range config.InterfaceRules {
		if rule.StructPattern == "" {
			return fmt.Errorf("interface rule %d: struct pattern cannot be empty", i)
		}
		if rule.InterfacePattern == "" {
			return fmt.Errorf("interface rule %d: interface pattern cannot be empty", i)
		}
		// Validate the patterns are valid regex
		if _, err := regexp.Compile(rule.StructPattern); err != nil {
			return fmt.Errorf("interface rule %d: invalid struct pattern: %w", i, err)
		}
		if _, err := regexp.Compile(rule.InterfacePattern); err != nil {
			return fmt.Errorf("interface rule %d: invalid interface pattern: %w", i, err)
		}
	}

	// Validate parameter rules
	for i, rule := range config.ParameterRules {
		if rule.StructPattern == "" {
			return fmt.Errorf("parameter rule %d: struct pattern cannot be empty", i)
		}
		if rule.MethodPattern == "" {
			return fmt.Errorf("parameter rule %d: method pattern cannot be empty", i)
		}
		if rule.ParameterTypePattern == "" {
			return fmt.Errorf("parameter rule %d: parameter type pattern cannot be empty", i)
		}
		// Validate the patterns are valid regex
		if _, err := regexp.Compile(rule.StructPattern); err != nil {
			return fmt.Errorf("parameter rule %d: invalid struct pattern: %w", i, err)
		}
		if _, err := regexp.Compile(rule.MethodPattern); err != nil {
			return fmt.Errorf("parameter rule %d: invalid method pattern: %w", i, err)
		}
		if _, err := regexp.Compile(rule.ParameterTypePattern); err != nil {
			return fmt.Errorf("parameter rule %d: invalid parameter type pattern: %w", i, err)
		}
	}

	// Validate layer-specific rules
	for i, rule := range config.LayerSpecificRules {
		if rule.Layer == "" {
			return fmt.Errorf("layer-specific rule %d: layer cannot be empty", i)
		}
		if !layerNames[rule.Layer] {
			return fmt.Errorf("layer-specific rule %d: references undefined layer: %s", i, rule.Layer)
		}
		if rule.RuleType == "" {
			return fmt.Errorf("layer-specific rule %d: rule type cannot be empty", i)
		}
		if rule.RuleType != "dependency" && rule.RuleType != "interface" && rule.RuleType != "parameter" {
			return fmt.Errorf("layer-specific rule %d: invalid rule type: %s", i, rule.RuleType)
		}
		if len(rule.Parameters) == 0 {
			return fmt.Errorf("layer-specific rule %d: parameters cannot be empty", i)
		}

		// Validate parameters based on rule type
		switch rule.RuleType {
		case "dependency":
			if _, ok := rule.Parameters["targetPattern"]; !ok {
				return fmt.Errorf("layer-specific rule %d: dependency rule requires 'targetPattern' parameter", i)
			}
			if _, err := regexp.Compile(rule.Parameters["targetPattern"]); err != nil {
				return fmt.Errorf("layer-specific rule %d: invalid target pattern: %w", i, err)
			}
		case "interface":
			if _, ok := rule.Parameters["structPattern"]; !ok {
				return fmt.Errorf("layer-specific rule %d: interface rule requires 'structPattern' parameter", i)
			}
			if _, ok := rule.Parameters["interfacePattern"]; !ok {
				return fmt.Errorf("layer-specific rule %d: interface rule requires 'interfacePattern' parameter", i)
			}
			if _, err := regexp.Compile(rule.Parameters["structPattern"]); err != nil {
				return fmt.Errorf("layer-specific rule %d: invalid struct pattern: %w", i, err)
			}
			if _, err := regexp.Compile(rule.Parameters["interfacePattern"]); err != nil {
				return fmt.Errorf("layer-specific rule %d: invalid interface pattern: %w", i, err)
			}
		case "parameter":
			if _, ok := rule.Parameters["structPattern"]; !ok {
				return fmt.Errorf("layer-specific rule %d: parameter rule requires 'structPattern' parameter", i)
			}
			if _, ok := rule.Parameters["methodPattern"]; !ok {
				return fmt.Errorf("layer-specific rule %d: parameter rule requires 'methodPattern' parameter", i)
			}
			if _, ok := rule.Parameters["parameterTypePattern"]; !ok {
				return fmt.Errorf("layer-specific rule %d: parameter rule requires 'parameterTypePattern' parameter", i)
			}
			if _, ok := rule.Parameters["shouldUseInterface"]; !ok {
				return fmt.Errorf("layer-specific rule %d: parameter rule requires 'shouldUseInterface' parameter", i)
			}
			if _, err := regexp.Compile(rule.Parameters["structPattern"]); err != nil {
				return fmt.Errorf("layer-specific rule %d: invalid struct pattern: %w", i, err)
			}
			if _, err := regexp.Compile(rule.Parameters["methodPattern"]); err != nil {
				return fmt.Errorf("layer-specific rule %d: invalid method pattern: %w", i, err)
			}
			if _, err := regexp.Compile(rule.Parameters["parameterTypePattern"]); err != nil {
				return fmt.Errorf("layer-specific rule %d: invalid parameter type pattern: %w", i, err)
			}
		}
	}

	// Validate direct layer dependency rules
	for i, rule := range config.DirectLayerDependencyRules {
		if rule.SourceLayer == "" {
			return fmt.Errorf("direct layer dependency rule %d: source layer cannot be empty", i)
		}
		if rule.TargetLayer == "" {
			return fmt.Errorf("direct layer dependency rule %d: target layer cannot be empty", i)
		}
		if !layerNames[rule.SourceLayer] {
			return fmt.Errorf("direct layer dependency rule %d: references undefined source layer: %s", i, rule.SourceLayer)
		}
		if !layerNames[rule.TargetLayer] {
			return fmt.Errorf("direct layer dependency rule %d: references undefined target layer: %s", i, rule.TargetLayer)
		}
	}

	return nil
}

// BuildArchitecture builds an architecture from the configuration
func (c *Config) BuildArchitecture(basePath string) (*arctest.Architecture, *arctest.LayeredArchitecture, []*arctest.DependencyRule, []*arctest.InterfaceImplementationRule, []*arctest.ParameterRule, error) {
	// Create a new architecture
	arch, err := arctest.New(basePath)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to create architecture: %w", err)
	}

	// Parse all packages
	if err := arch.ParsePackages(); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to parse packages: %w", err)
	}

	// Create layers
	layers := make([]*arctest.Layer, 0, len(c.Layers))
	layerMap := make(map[string]*arctest.Layer)

	for _, layerConfig := range c.Layers {
		layer, err := arctest.NewLayer(layerConfig.Name, layerConfig.Pattern)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to create layer %s: %w", layerConfig.Name, err)
		}
		layers = append(layers, layer)
		layerMap[layerConfig.Name] = layer
	}

	// Create layered architecture
	layeredArch := arch.NewLayeredArchitecture(layers...)

	// Add dependency rules from the basic rules section
	for _, ruleConfig := range c.Rules {
		fromLayer := layerMap[ruleConfig.From]
		toLayer := layerMap[ruleConfig.To]

		if err := fromLayer.DependsOnLayer(toLayer); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to add dependency rule from %s to %s: %w",
				ruleConfig.From, ruleConfig.To, err)
		}
	}

	// Create additional rule collections
	dependencyRules := make([]*arctest.DependencyRule, 0)
	interfaceRules := make([]*arctest.InterfaceImplementationRule, 0)
	parameterRules := make([]*arctest.ParameterRule, 0)

	// Add interface rules
	for _, ruleConfig := range c.InterfaceRules {
		rule, err := arch.StructsImplementInterfaces(ruleConfig.StructPattern, ruleConfig.InterfacePattern)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to create interface rule: %w", err)
		}
		interfaceRules = append(interfaceRules, rule)
	}

	// Add parameter rules
	for _, ruleConfig := range c.ParameterRules {
		var rule *arctest.ParameterRule
		var err error
		if ruleConfig.ShouldUseInterface {
			rule, err = arch.MethodsShouldUseInterfaceParameters(
				ruleConfig.StructPattern,
				ruleConfig.MethodPattern,
				ruleConfig.ParameterTypePattern,
			)
		} else {
			rule, err = arch.MethodsShouldUseStructParameters(
				ruleConfig.StructPattern,
				ruleConfig.MethodPattern,
				ruleConfig.ParameterTypePattern,
			)
		}
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to create parameter rule: %w", err)
		}
		parameterRules = append(parameterRules, rule)
	}

	// Add layer-specific rules
	for _, ruleConfig := range c.LayerSpecificRules {
		layer := layerMap[ruleConfig.Layer]
		if layer == nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("layer %s not found for layer-specific rule", ruleConfig.Layer)
		}

		switch ruleConfig.RuleType {
		case "dependency":
			targetPattern := ruleConfig.Parameters["targetPattern"]
			rule, err := layer.DoesNotDependOn(targetPattern)
			if err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to create layer-specific dependency rule: %w", err)
			}
			dependencyRules = append(dependencyRules, rule)
		case "interface":
			structPattern := ruleConfig.Parameters["structPattern"]
			interfacePattern := ruleConfig.Parameters["interfacePattern"]
			rule, err := layer.StructsImplementInterfaces(structPattern, interfacePattern)
			if err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to create layer-specific interface rule: %w", err)
			}
			interfaceRules = append(interfaceRules, rule)
		case "parameter":
			structPattern := ruleConfig.Parameters["structPattern"]
			methodPattern := ruleConfig.Parameters["methodPattern"]
			parameterTypePattern := ruleConfig.Parameters["parameterTypePattern"]
			shouldUseInterface := ruleConfig.Parameters["shouldUseInterface"] == "true"
			var rule *arctest.ParameterRule
			var err error
			if shouldUseInterface {
				rule, err = layer.MethodsShouldUseInterfaceParameters(
					structPattern,
					methodPattern,
					parameterTypePattern,
				)
			} else {
				rule, err = layer.MethodsShouldUseStructParameters(
					structPattern,
					methodPattern,
					parameterTypePattern,
				)
			}
			if err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to create layer-specific parameter rule: %w", err)
			}
			parameterRules = append(parameterRules, rule)
		}
	}

	// Add direct layer dependency rules
	for _, ruleConfig := range c.DirectLayerDependencyRules {
		sourceLayer := layerMap[ruleConfig.SourceLayer]
		targetLayer := layerMap[ruleConfig.TargetLayer]
		if sourceLayer == nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("source layer %s not found for direct layer dependency rule", ruleConfig.SourceLayer)
		}
		if targetLayer == nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("target layer %s not found for direct layer dependency rule", ruleConfig.TargetLayer)
		}

		var rule *arctest.DependencyRule
		var err error
		if ruleConfig.Allowed {
			// If allowed, add the dependency to the layered architecture
			if err := sourceLayer.DependsOnLayer(targetLayer); err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to add direct layer dependency rule: %w", err)
			}
		} else {
			// If not allowed, create a rule that the source layer should not depend on the target layer
			rule, err = sourceLayer.DoesNotDependOnLayer(targetLayer)
			if err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to create direct layer dependency rule: %w", err)
			}
			dependencyRules = append(dependencyRules, rule)
		}
	}

	return arch, layeredArch, dependencyRules, interfaceRules, parameterRules, nil
}

// RunArchitectureTests runs the architecture tests based on the configuration
func (c *Config) RunArchitectureTests(basePath string) (bool, []string, error) {
	arch, layeredArch, dependencyRules, interfaceRules, parameterRules, err := c.BuildArchitecture(basePath)
	if err != nil {
		return false, nil, err
	}

	allViolations := []string{}

	// Check layered architecture
	layerViolations, err := layeredArch.Check()
	if err != nil {
		return false, nil, fmt.Errorf("failed to check layered architecture: %w", err)
	}
	allViolations = append(allViolations, layerViolations...)

	// Check dependency rules
	if len(dependencyRules) > 0 {
		valid, violations := arch.ValidateDependenciesWithRules(dependencyRules)
		if !valid {
			allViolations = append(allViolations, violations...)
		}
	}

	// Check interface implementation rules
	if len(interfaceRules) > 0 {
		valid, violations := arch.ValidateInterfaceImplementations(interfaceRules)
		if !valid {
			allViolations = append(allViolations, violations...)
		}
	}

	// Check parameter rules
	if len(parameterRules) > 0 {
		valid, violations := arch.ValidateMethodParameters(parameterRules)
		if !valid {
			allViolations = append(allViolations, violations...)
		}
	}

	return len(allViolations) == 0, allViolations, nil
}

// SaveConfig saves the configuration to a YAML file
func (c *Config) SaveConfig(filePath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
