package examples

import (
	"testing"

	"github.com/mstrYoda/go-arctest/pkg/arctest"
)

func TestExampleProjectArchitecture(t *testing.T) {
	// Initialize architecture with project root
	arch, err := arctest.New("./example_project")
	if err != nil {
		t.Fatalf("Failed to create architecture: %v", err)
	}

	// Parse packages
	err = arch.ParsePackages("domain", "application", "infrastructure", "presentation")
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

	// Define layered architecture
	layeredArch := arch.NewLayeredArchitecture(
		domainLayer,
		applicationLayer,
		infrastructureLayer,
		presentationLayer,
	)

	// Define dependency rules
	// Domain layer should not depend on any other layer
	// Application layer may only depend on domain layer
	// Infrastructure layer may depend on domain layer
	// Presentation layer may depend on domain and application layers
	applicationLayer.DependsOnLayer(domainLayer)
	infrastructureLayer.DependsOnLayer(domainLayer)
	presentationLayer.DependsOnLayer(domainLayer)
	presentationLayer.DependsOnLayer(applicationLayer)

	// Check layered architecture
	violations, err := layeredArch.Check()
	if err != nil {
		t.Fatalf("Failed to check layered architecture: %v", err)
	}

	// Report any violations
	for _, violation := range violations {
		t.Errorf("Architecture violation: %s", violation)
	}

	// Test 1: Domain should not depend on any other layer
	rule1, err := arch.DoesNotDependOn("^domain$", "^(application|infrastructure|presentation)$")
	if err != nil {
		t.Fatalf("Failed to create dependency rule: %v", err)
	}

	// Test 2: Infrastructure repositories should implement domain interfaces
	rule2, err := arch.StructsImplementInterfaces(".*Repository$", ".*RepositoryInterface$")
	if err != nil {
		t.Fatalf("Failed to create interface implementation rule: %v", err)
	}

	// Test 3: Services should use interfaces as parameters, not concrete implementations
	rule3, err := arch.MethodsShouldUseInterfaceParameters(".*Service$", "New.*", ".*Repository$")
	if err != nil {
		t.Fatalf("Failed to create parameter rule: %v", err)
	}

	// Run the tests
	valid1, violations1 := arch.ValidateDependenciesWithRules([]*arctest.DependencyRule{rule1})
	if !valid1 {
		for _, violation := range violations1 {
			t.Errorf("Dependency violation: %s", violation)
		}
	}

	valid2, violations2 := arch.ValidateInterfaceImplementations([]*arctest.InterfaceImplementationRule{rule2})
	if !valid2 {
		for _, violation := range violations2 {
			t.Errorf("Interface implementation violation: %s", violation)
		}
	}

	valid3, violations3 := arch.ValidateMethodParameters([]*arctest.ParameterRule{rule3})
	if !valid3 {
		for _, violation := range violations3 {
			t.Errorf("Parameter type violation: %s", violation)
		}
	}
}
