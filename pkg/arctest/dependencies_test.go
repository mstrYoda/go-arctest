package arctest_test

import (
	"strings"
	"testing"

	"github.com/mstrYoda/go-arctest/pkg/arctest"
)

// TestDependencyRule_Creation tests creating dependency rules
func TestDependencyRule_Creation(t *testing.T) {
	t.Run("create valid rule", func(t *testing.T) {
		rule, err := arctest.NewDependencyRule("^domain$", "^infrastructure$", false)
		if err != nil {
			t.Fatalf("Failed to create dependency rule: %v", err)
		}
		if rule == nil {
			t.Fatal("Rule should not be nil")
		}
	})

	t.Run("with invalid regex patterns", func(t *testing.T) {
		_, err := arctest.NewDependencyRule("[invalid", "^infrastructure$", false)
		if err == nil {
			t.Fatal("Expected error for invalid source pattern, got nil")
		}

		_, err = arctest.NewDependencyRule("^domain$", "[invalid", false)
		if err == nil {
			t.Fatal("Expected error for invalid target pattern, got nil")
		}
	})
}

// TestArchitecture_CheckDependencies tests dependency checking
func TestArchitecture_CheckDependencies(t *testing.T) {
	arch, err := arctest.New("./testdata")
	if err != nil {
		t.Fatalf("Failed to create architecture: %v", err)
	}

	err = arch.ParsePackages()
	if err != nil {
		t.Fatalf("Failed to parse packages: %v", err)
	}

	t.Run("detect domain importing infrastructure violation", func(t *testing.T) {
		// Domain should not import infrastructure
		rule, err := arctest.NewDependencyRule("^domain$", ".*infrastructure.*", false)
		if err != nil {
			t.Fatalf("Failed to create dependency rule: %v", err)
		}

		violations, err := arch.CheckDependencies([]*arctest.DependencyRule{rule})
		if err != nil {
			t.Fatalf("Failed to check dependencies: %v", err)
		}

		// We expect a violation because domain/violation.go imports infrastructure
		if len(violations) == 0 {
			t.Fatal("Expected dependency violations, but found none")
		}

		foundViolation := false
		for _, v := range violations {
			if strings.Contains(v, "domain") && strings.Contains(v, "infrastructure") {
				foundViolation = true
				break
			}
		}
		if !foundViolation {
			t.Fatalf("Expected domain->infrastructure violation, but found: %v", violations)
		}
	})

	t.Run("allow infrastructure importing domain", func(t *testing.T) {
		// Infrastructure is allowed to import domain
		rule, err := arctest.NewDependencyRule("^infrastructure$", ".*domain.*", true)
		if err != nil {
			t.Fatalf("Failed to create dependency rule: %v", err)
		}

		violations, err := arch.CheckDependencies([]*arctest.DependencyRule{rule})
		if err != nil {
			t.Fatalf("Failed to check dependencies: %v", err)
		}

		// No violations expected
		if len(violations) > 0 {
			t.Fatalf("Expected no violations, but found: %v", violations)
		}
	})
}

// TestLayer_Creation tests creating Layer instances
func TestLayer_Creation(t *testing.T) {
	t.Run("create valid layer", func(t *testing.T) {
		layer, err := arctest.NewLayer("Domain", "^domain$")
		if err != nil {
			t.Fatalf("Failed to create layer: %v", err)
		}
		if layer == nil {
			t.Fatal("Layer should not be nil")
		}
		if layer.Name != "Domain" {
			t.Fatalf("Expected layer name 'Domain', got '%s'", layer.Name)
		}
	})

	t.Run("with multiple patterns", func(t *testing.T) {
		layer, err := arctest.NewLayer("Composite", "^domain$", "^application$")
		if err != nil {
			t.Fatalf("Failed to create layer with multiple patterns: %v", err)
		}
		if layer == nil {
			t.Fatal("Layer should not be nil")
		}
		if len(layer.Packages) != 2 {
			t.Fatalf("Expected 2 patterns, got %d", len(layer.Packages))
		}
	})

	t.Run("with invalid regex", func(t *testing.T) {
		_, err := arctest.NewLayer("Invalid", "[invalid")
		if err == nil {
			t.Fatal("Expected error for invalid regex, got nil")
		}
	})
}

// TestLayer_Contains tests layer containment
func TestLayer_Contains(t *testing.T) {
	layer, err := arctest.NewLayer("Domain", "^domain$")
	if err != nil {
		t.Fatalf("Failed to create layer: %v", err)
	}

	t.Run("matching package", func(t *testing.T) {
		if !layer.Contains("domain") {
			t.Fatal("Expected 'domain' to be contained in the layer")
		}
	})

	t.Run("non-matching package", func(t *testing.T) {
		if layer.Contains("application") {
			t.Fatal("Expected 'application' to not be contained in the layer")
		}
	})

	// Test with subpackage pattern
	layer, err = arctest.NewLayer("Domain", "domain")
	if err != nil {
		t.Fatalf("Failed to create layer: %v", err)
	}

	t.Run("matching subpackage", func(t *testing.T) {
		if !layer.Contains("domain/subpackage") {
			t.Fatal("Expected 'domain/subpackage' to be contained in the layer")
		}
	})
}

// TestLayeredArchitecture tests layered architecture functionality
func TestLayeredArchitecture(t *testing.T) {
	arch, err := arctest.New("./testdata")
	if err != nil {
		t.Fatalf("Failed to create architecture: %v", err)
	}

	err = arch.ParsePackages()
	if err != nil {
		t.Fatalf("Failed to parse packages: %v", err)
	}

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

	layeredArch := arch.NewLayeredArchitecture(
		domainLayer,
		applicationLayer,
		infrastructureLayer,
		presentationLayer,
	)

	t.Run("layer access via WhereLayer", func(t *testing.T) {
		layer := layeredArch.WhereLayer("Domain")
		if layer == nil {
			t.Fatal("Expected to find Domain layer")
		}
		if layer.Name != "Domain" {
			t.Fatalf("Expected layer name 'Domain', got '%s'", layer.Name)
		}

		nonExistentLayer := layeredArch.WhereLayer("NonExistent")
		if nonExistentLayer != nil {
			t.Fatal("Expected nil for non-existent layer")
		}
	})

	t.Run("define and check dependencies", func(t *testing.T) {
		// Define dependency rules
		err := applicationLayer.DependsOnLayer(domainLayer)
		if err != nil {
			t.Fatalf("Failed to define application->domain dependency: %v", err)
		}

		err = infrastructureLayer.DependsOnLayer(domainLayer)
		if err != nil {
			t.Fatalf("Failed to define infrastructure->domain dependency: %v", err)
		}

		err = presentationLayer.DependsOnLayer(domainLayer)
		if err != nil {
			t.Fatalf("Failed to define presentation->domain dependency: %v", err)
		}

		err = presentationLayer.DependsOnLayer(applicationLayer)
		if err != nil {
			t.Fatalf("Failed to define presentation->application dependency: %v", err)
		}

		// Check layered architecture
		violations, err := layeredArch.Check()
		if err != nil {
			t.Fatalf("Failed to check layered architecture: %v", err)
		}

		// We expect violations because domain imports infrastructure (in violation.go)
		domainViolation := false
		for _, v := range violations {
			if strings.Contains(v, "layer \"Domain\"") && strings.Contains(v, "layer \"Infrastructure\"") {
				domainViolation = true
				break
			}
		}

		if !domainViolation {
			t.Fatal("Expected violation from domain importing infrastructure")
		}
	})

	t.Run("does not depend on layer", func(t *testing.T) {
		rule, err := domainLayer.DoesNotDependOnLayer(infrastructureLayer)
		if err != nil {
			t.Fatalf("Failed to create DoesNotDependOnLayer rule: %v", err)
		}

		if rule == nil {
			t.Fatal("Expected rule to be created")
		}

		// Add rule to layered architecture
		layeredArch.AddDependencyConstraint(rule)

		// Test validating dependencies with rules
		valid, violations := arch.ValidateDependenciesWithRules([]*arctest.DependencyRule{rule})
		if valid {
			t.Fatal("Expected validation to fail due to domain->infrastructure dependency")
		}
		if len(violations) == 0 {
			t.Fatal("Expected violations, but found none")
		}
	})
}
