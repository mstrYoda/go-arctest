package examples

import (
	"testing"

	"github.com/mstrYoda/go-arctest/pkg/arctest"
)

// TestNestedPackages demonstrates how layer checks now handle nested packages
func TestNestedPackages(t *testing.T) {
	// Initialize architecture with project root
	arch, err := arctest.New("./example_project")
	if err != nil {
		t.Fatalf("Failed to create architecture: %v", err)
	}

	// Parse packages - this will now recursively discover and parse subpackages
	err = arch.ParsePackages("domain", "application")
	if err != nil {
		t.Fatalf("Failed to parse packages: %v", err)
	}

	// Define layers - these patterns will now match the main package and its subpackages
	domainLayer, err := arctest.NewLayer("Domain", "^domain$")
	if err != nil {
		t.Fatalf("Failed to create domain layer: %v", err)
	}

	applicationLayer, err := arctest.NewLayer("Application", "^application$")
	if err != nil {
		t.Fatalf("Failed to create application layer: %v", err)
	}

	// Print all discovered packages to verify subpackages were found
	t.Logf("Discovered packages:")
	for pkgPath := range arch.Packages {
		t.Logf("  - %s", pkgPath)

		// Check which layer contains this package
		if domainLayer.Contains(pkgPath) {
			t.Logf("    ✓ Package belongs to Domain layer")
		}
		if applicationLayer.Contains(pkgPath) {
			t.Logf("    ✓ Package belongs to Application layer")
		}
	}

	// Set up a layered architecture and set the architecture reference
	layeredArch := arch.NewLayeredArchitecture(domainLayer, applicationLayer)

	// Test layer dependency rule
	// Domain should not depend on Application layer (or any of its subpackages)
	domainAppRule, err := domainLayer.DoesNotDependOnLayer(applicationLayer)
	if err != nil {
		t.Fatalf("Failed to create layer dependency rule: %v", err)
	}

	// Validate the rule
	valid, violations := arch.ValidateDependenciesWithRules([]*arctest.DependencyRule{domainAppRule})
	if !valid {
		t.Logf("Dependency violations found (this is expected if domain imports from application):")
		for _, violation := range violations {
			t.Logf("  ✓ %s", violation)
		}
	} else {
		t.Logf("No dependency violations found - domain does not import from application or its subpackages")
	}

	// Only allow application to depend on domain
	err = applicationLayer.DependsOnLayer(domainLayer)
	if err != nil {
		t.Fatalf("Failed to create layer dependency: %v", err)
	}

	// Check the layered architecture - this should detect any violations including from/to subpackages
	layerViolations, err := layeredArch.Check()
	if err != nil {
		t.Fatalf("Failed to check layered architecture: %v", err)
	}

	if len(layerViolations) > 0 {
		t.Logf("Layered architecture violations found:")
		for _, violation := range layerViolations {
			t.Logf("  ✓ %s", violation)
		}
	} else {
		t.Logf("No layered architecture violations found")
	}
}

// TestNestedPackagesDependencyViolation demonstrates how layer checks now handle nested packages
func TestNestedPackagesDependencyViolation(t *testing.T) {
	// Initialize architecture with project root
	arch, err := arctest.New("./example_project")
	if err != nil {
		t.Fatalf("Failed to create architecture: %v", err)
	}

	// Parse packages - this will now recursively discover and parse subpackages
	err = arch.ParsePackages("domain", "application", "utils")
	if err != nil {
		t.Fatalf("Failed to parse packages: %v", err)
	}

	// Define layers - these patterns will now match the main package and its subpackages
	domainLayer, err := arctest.NewLayer("Domain", "^domain$")
	if err != nil {
		t.Fatalf("Failed to create domain layer: %v", err)
	}

	applicationLayer, err := arctest.NewLayer("Application", "^application$")
	if err != nil {
		t.Fatalf("Failed to create application layer: %v", err)
	}

	utilsLayer, err := arctest.NewLayer("Utils", "^utils$")
	if err != nil {
		t.Fatalf("Failed to create utils layer: %v", err)
	}

	// Print all discovered packages to verify subpackages were found
	t.Logf("Discovered packages:")
	for pkgPath := range arch.Packages {
		t.Logf("  - %s", pkgPath)

		// Check which layer contains this package
		if domainLayer.Contains(pkgPath) {
			t.Logf("    ✓ Package belongs to Domain layer")
		}
		if applicationLayer.Contains(pkgPath) {
			t.Logf("    ✓ Package belongs to Application layer")
		}
	}

	// Set up a layered architecture and set the architecture reference
	layeredArch := arch.NewLayeredArchitecture(domainLayer, applicationLayer, utilsLayer)

	// Test layer dependency rule
	// Domain should not depend on Application layer (or any of its subpackages)
	domainAppRule, err := domainLayer.DoesNotDependOnLayer(applicationLayer)
	if err != nil {
		t.Fatalf("Failed to create layer dependency rule: %v", err)
	}

	// Domain should not depend on Utils layer
	domainUtilsRule, err := domainLayer.DoesNotDependOnLayer(utilsLayer)
	if err != nil {
		t.Fatalf("Failed to create layer dependency rule: %v", err)
	}

	// Application should not depend on Utils layer
	applicationUtilsRule, err := applicationLayer.DoesNotDependOnLayer(utilsLayer)
	if err != nil {
		t.Fatalf("Failed to create layer dependency rule: %v", err)
	}

	// Validate the rules
	valid, violations := arch.ValidateDependenciesWithRules([]*arctest.DependencyRule{
		domainAppRule,        // Check domain -> app
		domainUtilsRule,      // Check domain -> utils
		applicationUtilsRule, // Check app -> utils
	})

	if !valid {
		t.Logf("Dependency violations found (expected):")
		for _, violation := range violations {
			t.Logf("  ✓ %s", violation)
		}
	} else {
		t.Error("Expected dependency violations, but none were found!")
	}

	// Only allow application to depend on domain
	err = applicationLayer.DependsOnLayer(domainLayer)
	if err != nil {
		t.Fatalf("Failed to create layer dependency: %v", err)
	}

	// Check the layered architecture - this should detect any violations including from/to subpackages
	layerViolations, err := layeredArch.Check()
	if err != nil {
		t.Fatalf("Failed to check layered architecture: %v", err)
	}

	if len(layerViolations) > 0 {
		t.Logf("Layered architecture violations found:")
		for _, violation := range layerViolations {
			t.Logf("  ✓ %s", violation)
		}
	} else {
		t.Logf("No layered architecture violations found")
	}
}
