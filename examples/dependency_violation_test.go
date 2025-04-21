package examples

import (
	"testing"

	"github.com/mstrYoda/go-arctest/pkg/arctest"
)

// TestDependencyViolation demonstrates how architecture tests detect dependency violations
func TestDependencyViolation(t *testing.T) {
	// Initialize architecture with project root
	arch, err := arctest.New("./example_project")
	if err != nil {
		t.Fatalf("Failed to create architecture: %v", err)
	}

	// Parse packages
	err = arch.ParsePackages("domain", "application", "infrastructure", "presentation", "utils")
	if err != nil {
		t.Fatalf("Failed to parse packages: %v", err)
	}

	// Define layers
	domainLayer, err := arctest.NewLayer("Domain", "^domain$")
	if err != nil {
		t.Fatalf("Failed to create domain layer: %v", err)
	}

	utilsLayer, err := arctest.NewLayer("Utils", "^utils$")
	if err != nil {
		t.Fatalf("Failed to create utils layer: %v", err)
	}

	// Define layered architecture
	layeredArch := arctest.NewLayeredArchitecture(
		domainLayer,
		utilsLayer,
	)

	// Set architecture for the layered architecture
	layeredArch.SetArchitecture(arch)

	// Create a rule that domain should not depend on utils
	// We need to match the full import path, not just the "utils" part
	// Using ".*utils$" to match any import that ends with "utils"
	rule, err := domainLayer.DoesNotDependOn(".*utils$")
	if err != nil {
		t.Fatalf("Failed to create dependency rule: %v", err)
	}

	// Run the test
	valid, violations := arch.ValidateDependenciesWithRules([]*arctest.DependencyRule{rule})

	// For a standard test, we would assert that valid is true.
	// But since we intentionally created a violation, we expect valid to be false.
	if valid {
		t.Error("Expected dependency violations, but none were found!")
	} else {
		t.Logf("Successfully detected dependency violations:")
		for _, violation := range violations {
			t.Logf("  ✓ %s", violation)
		}
	}
}

// TestDependencyViolationWithLayers demonstrates how to use layer-specific rules to detect violations
func TestDependencyViolationWithLayers(t *testing.T) {
	// Initialize architecture with project root
	arch, err := arctest.New("./example_project")
	if err != nil {
		t.Fatalf("Failed to create architecture: %v", err)
	}

	// Parse packages
	err = arch.ParsePackages("domain", "application", "infrastructure", "presentation", "utils")
	if err != nil {
		t.Fatalf("Failed to parse packages: %v", err)
	}

	// Define layers
	domainLayer, err := arctest.NewLayer("Domain", "^domain$")
	if err != nil {
		t.Fatalf("Failed to create domain layer: %v", err)
	}

	applicationLayer, err := arctest.NewLayer("Application", "^application$")
	if err != nil {
		t.Fatalf("Failed to create application layer: %v", err)
	}

	infrastructureLayer, err := arctest.NewLayer("Infrastructure", "^infrastructure$")
	if err != nil {
		t.Fatalf("Failed to create infrastructure layer: %v", err)
	}

	presentationLayer, err := arctest.NewLayer("Presentation", "^presentation$")
	if err != nil {
		t.Fatalf("Failed to create presentation layer: %v", err)
	}

	utilsLayer, err := arctest.NewLayer("Utils", "^utils$")
	if err != nil {
		t.Fatalf("Failed to create utils layer: %v", err)
	}

	// Define layered architecture
	layeredArch := arctest.NewLayeredArchitecture(
		domainLayer,
		applicationLayer,
		infrastructureLayer,
		presentationLayer,
		utilsLayer,
	)

	// Set architecture for the layered architecture
	layeredArch.SetArchitecture(arch)

	// Define allowed dependencies
	applicationLayer.DependsOn("Domain", layeredArch)
	applicationLayer.DependsOn("Utils", layeredArch)
	infrastructureLayer.DependsOn("Domain", layeredArch)
	infrastructureLayer.DependsOn("Utils", layeredArch)
	presentationLayer.DependsOn("Domain", layeredArch)
	presentationLayer.DependsOn("Application", layeredArch)
	presentationLayer.DependsOn("Utils", layeredArch)

	// Intentionally NOT allowing Domain to depend on Utils

	// Check layered architecture
	violations, err := layeredArch.Check(arch)
	if err != nil {
		t.Fatalf("Failed to check layered architecture: %v", err)
	}

	// We expect violations because domain is importing utils, but we didn't define a rule allowing it
	if len(violations) == 0 {
		t.Error("Expected dependency violations in layered architecture, but none were found!")
	} else {
		t.Logf("Successfully detected dependency violations in layered architecture:")
		for _, violation := range violations {
			t.Logf("  ✓ %s", violation)
		}
	}
}
