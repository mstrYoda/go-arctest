package examples

import (
	"testing"

	"github.com/mstrYoda/go-arctest/pkg/arctest"
)

// TestInterfaceParameterViolation demonstrates how to test for violations
// where a method should use an interface parameter but uses a concrete struct instead
func TestInterfaceParameterViolation(t *testing.T) {
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

	// First, let's modify the user_with_dependency_violation.go file to add a method with a parameter
	// that violates our architectural rule

	// Now, create a rule that methods in UserServiceWithLogger should use the Logger interface
	// as a parameter, but we know it uses the concrete *utils.Logger struct
	rule, err := arch.MethodsShouldUseStructParameters(".*Service.*", ".*", ".*Logger")
	if err != nil {
		t.Fatalf("Failed to create parameter rule: %v", err)
	}

	// Validate the parameter usage - we expect this to pass because the struct DOES use concrete Logger
	valid, violations := arch.ValidateMethodParameters([]*arctest.ParameterRule{rule})

	// We expect this validation to pass since our rule is looking for struct parameters and that's what we have
	if !valid {
		t.Error("Unexpected parameter type violations found!")
		for _, violation := range violations {
			t.Errorf("  ✗ %s", violation)
		}
	} else {
		t.Logf("Successfully validated that concrete struct parameters are used (intentional violation of good practice)")
	}

	// Now let's test the inverse rule - methods should use Logger interfaces, not structs
	interfaceRule, err := arch.MethodsShouldUseInterfaceParameters(".*Service.*", ".*", ".*Logger")
	if err != nil {
		t.Fatalf("Failed to create interface parameter rule: %v", err)
	}

	// This should detect violations since we're using concrete structs
	interfaceValid, interfaceViolations := arch.ValidateMethodParameters([]*arctest.ParameterRule{interfaceRule})

	// We expect this to fail - violation should be detected
	if interfaceValid {
		t.Error("Expected interface parameter violations, but none were found!")
	} else {
		t.Logf("Successfully detected interface parameter violations (methods using concrete types instead of interfaces):")
		for _, violation := range interfaceViolations {
			t.Logf("  ✓ %s", violation)
		}
	}
}
