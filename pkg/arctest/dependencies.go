package arctest

import (
	"fmt"
	"regexp"
	"strings"
)

// DependencyRule represents a dependency rule
type DependencyRule struct {
	SourcePattern      string // regex pattern for source package
	TargetPattern      string // regex pattern for target package
	AllowedImports     bool   // if true, source can import target, if false, source cannot import target
	sourcePatternRegex *regexp.Regexp
	targetPatternRegex *regexp.Regexp
}

// NewDependencyRule creates a new dependency rule
func NewDependencyRule(sourcePattern, targetPattern string, allowedImports bool) (*DependencyRule, error) {
	sourceRegex, err := regexp.Compile(sourcePattern)
	if err != nil {
		return nil, fmt.Errorf("invalid source pattern: %w", err)
	}

	targetRegex, err := regexp.Compile(targetPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid target pattern: %w", err)
	}

	return &DependencyRule{
		SourcePattern:      sourcePattern,
		TargetPattern:      targetPattern,
		AllowedImports:     allowedImports,
		sourcePatternRegex: sourceRegex,
		targetPatternRegex: targetRegex,
	}, nil
}

// CheckDependencies checks all packages against the provided dependency rules
func (a *Architecture) CheckDependencies(rules []*DependencyRule) ([]string, error) {
	violations := []string{}

	for pkgPath, pkg := range a.Packages {
		for _, importPath := range pkg.Imports {
			for _, rule := range rules {
				// Check if this package matches the source pattern
				if rule.sourcePatternRegex.MatchString(pkgPath) {
					// Check if the import matches the target pattern
					if rule.targetPatternRegex.MatchString(importPath) {
						// If imports are not allowed, this is a violation
						if !rule.AllowedImports {
							violations = append(violations, fmt.Sprintf(
								"Package %q imports %q, but this is not allowed by rule: %s cannot import %s",
								pkgPath, importPath, rule.SourcePattern, rule.TargetPattern,
							))
						}
					}
				}
			}
		}
	}

	return violations, nil
}

// Layer represents a layer in a layered architecture
type Layer struct {
	Name     string
	Packages []string // Package paths or patterns
	patterns []*regexp.Regexp
	arch     *Architecture // Reference to the architecture
}

// NewLayer creates a new layer with the given name and package patterns
func NewLayer(name string, packages ...string) (*Layer, error) {
	patterns := make([]*regexp.Regexp, 0, len(packages))

	for _, pkg := range packages {
		pattern, err := regexp.Compile(pkg)
		if err != nil {
			return nil, fmt.Errorf("invalid package pattern %q: %w", pkg, err)
		}
		patterns = append(patterns, pattern)
	}

	return &Layer{
		Name:     name,
		Packages: packages,
		patterns: patterns,
	}, nil
}

// Contains checks if a package belongs to this layer
func (l *Layer) Contains(pkgPath string) bool {
	for _, pattern := range l.patterns {
		if pattern.MatchString(pkgPath) {
			return true
		}
	}
	return false
}

// SetArchitecture sets the architecture reference for this layer
// This is called internally when the layer is added to a layered architecture
func (l *Layer) SetArchitecture(arch *Architecture) {
	l.arch = arch
}

// DependsOn creates a rule that this layer may depend on another layer
func (l *Layer) DependsOn(targetLayerName string, layeredArch *LayeredArchitecture) error {
	return layeredArch.AddRule(l.Name, targetLayerName)
}

// DependsOnLayer creates a rule that this layer may depend on another layer directly
func (l *Layer) DependsOnLayer(targetLayer *Layer, layeredArch *LayeredArchitecture) error {
	if targetLayer == nil {
		return fmt.Errorf("target layer cannot be nil")
	}
	return layeredArch.AddRule(l.Name, targetLayer.Name)
}

// DoesNotDependOn creates a rule that this layer should not depend on a specific package pattern
func (l *Layer) DoesNotDependOn(targetPattern string) (*DependencyRule, error) {
	if l.arch == nil {
		return nil, fmt.Errorf("layer %q is not associated with an architecture", l.Name)
	}

	// Create patterns that will match any fully qualified import path
	// ending with the specified package patterns
	sourcePatterns := make([]string, 0, len(l.Packages))
	for _, pkg := range l.Packages {
		// Remove ^ and $ markers if present
		cleanPattern := strings.TrimPrefix(pkg, "^")
		cleanPattern = strings.TrimSuffix(cleanPattern, "$")
		// Create pattern that matches any path ending with the package
		sourcePatterns = append(sourcePatterns, fmt.Sprintf("(^|/)%s$", cleanPattern))
	}
	sourcePattern := strings.Join(sourcePatterns, "|")

	return NewDependencyRule(sourcePattern, targetPattern, false)
}

// DoesNotDependOnLayer creates a rule that this layer should not depend on another layer
func (l *Layer) DoesNotDependOnLayer(targetLayer *Layer) (*DependencyRule, error) {
	if l.arch == nil {
		return nil, fmt.Errorf("layer %q is not associated with an architecture", l.Name)
	}

	if targetLayer == nil {
		return nil, fmt.Errorf("target layer cannot be nil")
	}

	// Create patterns that will match any fully qualified import path
	// ending with the specified package patterns
	sourcePatterns := make([]string, 0, len(l.Packages))
	for _, pkg := range l.Packages {
		// Remove ^ and $ markers if present
		cleanPattern := strings.TrimPrefix(pkg, "^")
		cleanPattern = strings.TrimSuffix(cleanPattern, "$")
		// Create pattern that matches any path ending with the package
		sourcePatterns = append(sourcePatterns, fmt.Sprintf("(^|/)%s$", cleanPattern))
	}
	sourcePattern := strings.Join(sourcePatterns, "|")

	// Same for target patterns
	targetPatterns := make([]string, 0, len(targetLayer.Packages))
	for _, pkg := range targetLayer.Packages {
		cleanPattern := strings.TrimPrefix(pkg, "^")
		cleanPattern = strings.TrimSuffix(cleanPattern, "$")
		targetPatterns = append(targetPatterns, fmt.Sprintf("(^|/)%s$", cleanPattern))
	}
	targetPattern := strings.Join(targetPatterns, "|")

	// Create a rule that disallows dependencies from source to target
	rule, err := NewDependencyRule(sourcePattern, targetPattern, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create dependency rule: %w", err)
	}

	return rule, nil
}

// StructsImplementInterfaces creates a rule that structs in this layer matching a pattern
// must implement interfaces matching a pattern
func (l *Layer) StructsImplementInterfaces(structPattern, interfacePattern string) (*InterfaceImplementationRule, error) {
	if l.arch == nil {
		return nil, fmt.Errorf("layer %q is not associated with an architecture", l.Name)
	}

	// Modify the struct pattern to only match structs in this layer's packages
	layerScopedStructPattern := l.getScopedPattern(structPattern)
	return NewInterfaceImplementationRule(layerScopedStructPattern, interfacePattern)
}

// MethodsShouldUseInterfaceParameters creates a rule that methods in this layer should use interface parameters
func (l *Layer) MethodsShouldUseInterfaceParameters(structPattern, methodPattern, parameterTypePattern string) (*ParameterRule, error) {
	if l.arch == nil {
		return nil, fmt.Errorf("layer %q is not associated with an architecture", l.Name)
	}

	// Modify the struct pattern to only match structs in this layer's packages
	layerScopedStructPattern := l.getScopedPattern(structPattern)
	return NewParameterRule(layerScopedStructPattern, methodPattern, parameterTypePattern, true)
}

// MethodsShouldUseStructParameters creates a rule that methods in this layer should use struct parameters
func (l *Layer) MethodsShouldUseStructParameters(structPattern, methodPattern, parameterTypePattern string) (*ParameterRule, error) {
	if l.arch == nil {
		return nil, fmt.Errorf("layer %q is not associated with an architecture", l.Name)
	}

	// Modify the struct pattern to only match structs in this layer's packages
	layerScopedStructPattern := l.getScopedPattern(structPattern)
	return NewParameterRule(layerScopedStructPattern, methodPattern, parameterTypePattern, false)
}

// getScopedPattern prefixes the pattern with the layer's package patterns
func (l *Layer) getScopedPattern(pattern string) string {
	// If the pattern is already scoped to packages, leave it as is
	if strings.HasPrefix(pattern, "^") {
		return pattern
	}

	// Simple approach: prefix the pattern with each package pattern
	var scopedPatterns []string
	for _, pkgPattern := range l.Packages {
		// Remove ^ and $ from package pattern if present
		pkgPattern = strings.TrimPrefix(pkgPattern, "^")
		pkgPattern = strings.TrimSuffix(pkgPattern, "$")

		// Combine the package pattern with the input pattern
		scopedPatterns = append(scopedPatterns, fmt.Sprintf("^%s\\.%s", pkgPattern, pattern))
	}

	return "(" + strings.Join(scopedPatterns, "|") + ")"
}

// LayeredArchitecture represents a layered architecture with dependency rules
type LayeredArchitecture struct {
	Layers [](*Layer)
	rules  [](*DependencyRule)
	arch   *Architecture // Reference to the architecture
}

// NewLayeredArchitecture creates a new layered architecture
func NewLayeredArchitecture(layers ...*Layer) *LayeredArchitecture {
	return &LayeredArchitecture{
		Layers: layers,
		rules:  make([]*DependencyRule, 0),
	}
}

// SetArchitecture sets the architecture reference for this layered architecture and all its layers
func (la *LayeredArchitecture) SetArchitecture(arch *Architecture) {
	la.arch = arch
	for _, layer := range la.Layers {
		layer.SetArchitecture(arch)
	}
}

// WhereLayer returns a layer by name
func (la *LayeredArchitecture) WhereLayer(name string) *Layer {
	for _, layer := range la.Layers {
		if layer.Name == name {
			return layer
		}
	}
	return nil
}

// AddRule adds a rule that layer A may import layer B
func (la *LayeredArchitecture) AddRule(sourceLayerName, targetLayerName string) error {
	sourceLayer := la.WhereLayer(sourceLayerName)
	if sourceLayer == nil {
		return fmt.Errorf("source layer %q not found", sourceLayerName)
	}

	targetLayer := la.WhereLayer(targetLayerName)
	if targetLayer == nil {
		return fmt.Errorf("target layer %q not found", targetLayerName)
	}

	// Create patterns for all packages in source and target layers
	for _, sourcePkg := range sourceLayer.Packages {
		for _, targetPkg := range targetLayer.Packages {
			// Clean up patterns by removing ^ and $ if present
			sourceClean := strings.TrimPrefix(sourcePkg, "^")
			sourceClean = strings.TrimSuffix(sourceClean, "$")
			targetClean := strings.TrimPrefix(targetPkg, "^")
			targetClean = strings.TrimSuffix(targetClean, "$")

			// Create patterns that match any path ending with the package name
			sourcePattern := fmt.Sprintf("(^|/)%s$", sourceClean)
			targetPattern := fmt.Sprintf("(^|/)%s$", targetClean)

			rule, err := NewDependencyRule(sourcePattern, targetPattern, true)
			if err != nil {
				return err
			}
			la.rules = append(la.rules, rule)
		}
	}

	return nil
}

// AddDependencyConstraint adds a dependency constraint rule directly to the layered architecture
func (la *LayeredArchitecture) AddDependencyConstraint(rule *DependencyRule) {
	la.rules = append(la.rules, rule)
}

// Check checks the architecture against the defined layers and rules
func (la *LayeredArchitecture) Check(arch *Architecture) ([]string, error) {
	// Set the architecture reference
	la.SetArchitecture(arch)

	violations := []string{}

	// For each package, check which layer it belongs to
	for pkgPath, pkg := range arch.Packages {
		var sourceLayer *Layer
		for _, layer := range la.Layers {
			if layer.Contains(pkgPath) {
				sourceLayer = layer
				break
			}
		}

		if sourceLayer == nil {
			// Skip packages that don't belong to any layer
			continue
		}

		// Check each import
		for _, importPath := range pkg.Imports {
			// Skip standard library imports that don't have dots or slashes
			// (this generally means they're from the standard library)
			if !strings.Contains(importPath, ".") && !strings.Contains(importPath, "/") {
				continue
			}

			// Find which layer the import belongs to
			var targetLayer *Layer
			for _, layer := range la.Layers {
				// Check if this import belongs to the layer
				for _, pattern := range layer.patterns {
					// Improve matching to detect the layer based on the import path
					// For packages like github.com/mstrYoda/go-arctest/examples/example_project/utils
					// we want to match against the "utils" part
					if pattern.MatchString(importPath) ||
						strings.HasSuffix(importPath, "/"+strings.TrimPrefix(strings.TrimSuffix(layer.Packages[0], "$"), "^")) {
						targetLayer = layer
						break
					}
				}
				if targetLayer != nil {
					break
				}
			}

			if targetLayer == nil {
				// Skip imports that don't belong to any layer
				continue
			}

			// Skip if it's the same layer
			if sourceLayer == targetLayer {
				continue
			}

			// Check if this import is allowed by rules
			allowed := false
			for _, rule := range la.rules {
				if rule.sourcePatternRegex.MatchString(pkgPath) &&
					rule.targetPatternRegex.MatchString(importPath) &&
					rule.AllowedImports {
					allowed = true
					break
				}
			}

			if !allowed {
				violations = append(violations, fmt.Sprintf(
					"Package %q in layer %q imports %q in layer %q, but no rule allows this dependency",
					pkgPath, sourceLayer.Name, importPath, targetLayer.Name,
				))
			}
		}
	}

	return violations, nil
}

// DependsOn creates a rule that one package pattern depends on another
func (a *Architecture) DependsOn(sourcePattern, targetPattern string) (*DependencyRule, error) {
	return NewDependencyRule(sourcePattern, targetPattern, true)
}

// DoesNotDependOn creates a rule that one package pattern should not depend on another
func (a *Architecture) DoesNotDependOn(sourcePattern, targetPattern string) (*DependencyRule, error) {
	return NewDependencyRule(sourcePattern, targetPattern, false)
}

// ValidateDependenciesWithRules validates dependencies against the provided rules
func (a *Architecture) ValidateDependenciesWithRules(rules []*DependencyRule) (bool, []string) {
	violations, _ := a.CheckDependencies(rules)
	return len(violations) == 0, violations
}
