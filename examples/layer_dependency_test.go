package examples

import (
	"testing"

	"github.com/mstrYoda/go-arctest/pkg/arctest"
)

// TestLayerDependencyRules demonstrates how to use direct layer dependency rules
func TestLayerDependencyRules(t *testing.T) {
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

	// Method 1: Define direct layer dependencies using the new API
	// This is more intuitive and easier to read than using regex patterns

	// Domain should not depend on Application layer
	domainAppRule, err := domainLayer.DoesNotDependOnLayer(applicationLayer)
	if err != nil {
		t.Fatalf("Failed to create layer dependency rule: %v", err)
	}

	// Domain should not depend on Utils layer (we know this is violated in the example)
	domainUtilsRule, err := domainLayer.DoesNotDependOnLayer(utilsLayer)
	if err != nil {
		t.Fatalf("Failed to create layer dependency rule: %v", err)
	}

	// Test the rules
	valid, violations := arch.ValidateDependenciesWithRules([]*arctest.DependencyRule{
		domainAppRule,
		domainUtilsRule,
	})

	// We expect a violation in the domain -> utils dependency
	if valid {
		t.Error("Expected dependency violations, but none were found!")
	} else {
		t.Logf("Successfully detected dependency violations:")
		for _, violation := range violations {
			t.Logf("  ✓ %s", violation)
		}
	}

	// Method 2: Using the dependency layer in a LayeredArchitecture
	// Set up allowed dependencies with the new API
	err = applicationLayer.DependsOnLayer(domainLayer)
	if err != nil {
		t.Fatalf("Failed to create layer dependency: %v", err)
	}

	err = applicationLayer.DependsOnLayer(utilsLayer)
	if err != nil {
		t.Fatalf("Failed to create layer dependency: %v", err)
	}

	err = infrastructureLayer.DependsOnLayer(domainLayer)
	if err != nil {
		t.Fatalf("Failed to create layer dependency: %v", err)
	}

	err = infrastructureLayer.DependsOnLayer(utilsLayer)
	if err != nil {
		t.Fatalf("Failed to create layer dependency: %v", err)
	}

	err = presentationLayer.DependsOnLayer(domainLayer)
	if err != nil {
		t.Fatalf("Failed to create layer dependency: %v", err)
	}

	err = presentationLayer.DependsOnLayer(applicationLayer)
	if err != nil {
		t.Fatalf("Failed to create layer dependency: %v", err)
	}

	// Unlike the string-based API, we don't need to add a utils dependency

	// Check layered architecture
	layerViolations, err := layeredArch.Check()
	if err != nil {
		t.Fatalf("Failed to check layered architecture: %v", err)
	}

	// We expect a violation because domain is importing utils, but we didn't define a rule allowing it
	if len(layerViolations) == 0 {
		t.Error("Expected dependency violations in layered architecture, but none were found!")
	} else {
		t.Logf("Successfully detected dependency violations in layered architecture:")
		for _, violation := range layerViolations {
			t.Logf("  ✓ %s", violation)
		}
	}
}
