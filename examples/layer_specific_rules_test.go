package examples

import (
	"testing"

	"github.com/mstrYoda/go-arctest/pkg/arctest"
)

// Example architecture test using layer-specific rules
func TestLayerSpecificRules(t *testing.T) {
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
	layeredArch := arctest.NewLayeredArchitecture(
		domainLayer,
		applicationLayer,
		infrastructureLayer,
		presentationLayer,
	)

	// Set architecture for the layered architecture
	layeredArch.SetArchitecture(arch)

	// Define layer dependencies using layer-specific methods
	if err := applicationLayer.DependsOn("Domain", layeredArch); err != nil {
		t.Fatalf("Failed to define layer dependency: %v", err)
	}

	if err := infrastructureLayer.DependsOn("Domain", layeredArch); err != nil {
		t.Fatalf("Failed to define layer dependency: %v", err)
	}

	if err := presentationLayer.DependsOn("Domain", layeredArch); err != nil {
		t.Fatalf("Failed to define layer dependency: %v", err)
	}

	if err := presentationLayer.DependsOn("Application", layeredArch); err != nil {
		t.Fatalf("Failed to define layer dependency: %v", err)
	}

	// Check layered architecture
	violations, err := layeredArch.Check(arch)
	if err != nil {
		t.Fatalf("Failed to check layered architecture: %v", err)
	}

	// Report any violations
	for _, violation := range violations {
		t.Errorf("Architecture violation: %s", violation)
	}

	// Layer-specific rules:

	// 1. Domain layer should not depend on any other layer
	rule1, err := domainLayer.DoesNotDependOn("^(application|infrastructure|presentation)$")
	if err != nil {
		t.Fatalf("Failed to create domain layer dependency rule: %v", err)
	}

	// 2. Infrastructure repositories should implement domain interfaces
	rule2, err := infrastructureLayer.StructsImplementInterfaces(".*Repository$", ".*RepositoryInterface$")
	if err != nil {
		t.Fatalf("Failed to create interface implementation rule: %v", err)
	}

	// 3. Application services should use interfaces as parameters
	rule3, err := applicationLayer.MethodsShouldUseInterfaceParameters(".*Service$", "New.*", ".*Repository$")
	if err != nil {
		t.Fatalf("Failed to create parameter rule: %v", err)
	}

	// 4. Presentation handlers should use application services
	rule4, err := presentationLayer.MethodsShouldUseInterfaceParameters(".*Handler$", "New.*", ".*Service$")
	if err != nil {
		t.Fatalf("Failed to create parameter rule: %v", err)
	}

	// Run the tests
	valid1, violations1 := arch.ValidateDependenciesWithRules([]*arctest.DependencyRule{rule1})
	if !valid1 {
		for _, violation := range violations1 {
			t.Errorf("Domain layer violation: %s", violation)
		}
	}

	valid2, violations2 := arch.ValidateInterfaceImplementations([]*arctest.InterfaceImplementationRule{rule2})
	if !valid2 {
		for _, violation := range violations2 {
			t.Errorf("Repository implementation violation: %s", violation)
		}
	}

	valid3, violations3 := arch.ValidateMethodParameters([]*arctest.ParameterRule{rule3})
	if !valid3 {
		for _, violation := range violations3 {
			t.Errorf("Service parameter violation: %s", violation)
		}
	}

	valid4, violations4 := arch.ValidateMethodParameters([]*arctest.ParameterRule{rule4})
	if !valid4 {
		for _, violation := range violations4 {
			t.Errorf("Handler parameter violation: %s", violation)
		}
	}
}
