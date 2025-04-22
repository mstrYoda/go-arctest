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
	layeredArch := arch.NewLayeredArchitecture(
		domainLayer,
		utilsLayer,
	)

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

	violations, err = layeredArch.Check()
	if err != nil {
		t.Fatalf("Failed to check layered architecture: %v", err)
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
	layeredArch := arch.NewLayeredArchitecture(
		domainLayer,
		applicationLayer,
		infrastructureLayer,
		presentationLayer,
		utilsLayer,
	)

	// Define allowed dependencies
	applicationLayer.DependsOn("Domain")
	applicationLayer.DependsOn("Utils")
	infrastructureLayer.DependsOn("Domain")
	infrastructureLayer.DependsOn("Utils")
	presentationLayer.DependsOn("Domain")
	presentationLayer.DependsOn("Application")
	presentationLayer.DependsOn("Utils")

	// Intentionally NOT allowing Domain to depend on Utils

	// Check layered architecture
	violations, err := layeredArch.Check()
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
