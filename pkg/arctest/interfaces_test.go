package arctest_test

import (
	"testing"

	"github.com/mstrYoda/go-arctest/pkg/arctest"
)

// TestInterfaceImplementationRule tests creation of interface rules
func TestInterfaceImplementationRule(t *testing.T) {
	t.Run("create valid rule", func(t *testing.T) {
		rule, err := arctest.NewInterfaceImplementationRule(".*Repository$", ".*RepositoryInterface$")
		if err != nil {
			t.Fatalf("Failed to create interface rule: %v", err)
		}
		if rule == nil {
			t.Fatal("Rule should not be nil")
		}
	})

	t.Run("with invalid regex patterns", func(t *testing.T) {
		_, err := arctest.NewInterfaceImplementationRule("[invalid", ".*RepositoryInterface$")
		if err == nil {
			t.Fatal("Expected error for invalid struct pattern, got nil")
		}

		_, err = arctest.NewInterfaceImplementationRule(".*Repository$", "[invalid")
		if err == nil {
			t.Fatal("Expected error for invalid interface pattern, got nil")
		}
	})
}

// TestArchitecture_CheckStructImplementsInterfaces tests interface implementation checks
func TestArchitecture_CheckStructImplementsInterfaces(t *testing.T) {
	arch, err := arctest.New("./testdata")
	if err != nil {
		t.Fatalf("Failed to create architecture: %v", err)
	}

	err = arch.ParsePackages()
	if err != nil {
		t.Fatalf("Failed to parse packages: %v", err)
	}

	t.Run("detect missing implementation", func(t *testing.T) {
		// Check that repositories implement repository interfaces
		rule, err := arctest.NewInterfaceImplementationRule(".*Repository$", ".*RepositoryInterface$")
		if err != nil {
			t.Fatalf("Failed to create interface rule: %v", err)
		}

		violations, err := arch.CheckStructImplementsInterfaces([]*arctest.InterfaceImplementationRule{rule})
		if err != nil {
			t.Fatalf("Failed to check interface implementations: %v", err)
		}

		// We expect a violation for ProductRepository which doesn't implement ProductRepositoryInterface
		// because it's missing the Save method
		if len(violations) == 0 {
			t.Fatal("Expected interface implementation violations, but found none")
		}

		foundViolation := false
		for _, v := range violations {
			if v == "Struct \"ProductRepository\" in package \"infrastructure\" does not implement any interface matching \".*RepositoryInterface$\"" {
				foundViolation = true
				break
			}
		}

		if !foundViolation {
			t.Fatalf("Expected ProductRepository violation, but found: %v", violations)
		}
	})

	t.Run("successful implementation", func(t *testing.T) {
		// UserRepository should implement UserRepositoryInterface
		// Create a more specific rule just for UserRepository
		rule, err := arctest.NewInterfaceImplementationRule("UserRepository$", "UserRepositoryInterface$")
		if err != nil {
			t.Fatalf("Failed to create interface rule: %v", err)
		}

		violations, err := arch.CheckStructImplementsInterfaces([]*arctest.InterfaceImplementationRule{rule})
		if err != nil {
			t.Fatalf("Failed to check interface implementations: %v", err)
		}

		// No violations expected for UserRepository
		if len(violations) > 0 {
			t.Fatalf("Expected no violations for UserRepository, but found: %v", violations)
		}
	})
}

// TestArchitecture_StructsImplementInterfaces tests the helper method for creating rules
func TestArchitecture_StructsImplementInterfaces(t *testing.T) {
	arch, err := arctest.New("./testdata")
	if err != nil {
		t.Fatalf("Failed to create architecture: %v", err)
	}

	t.Run("create rule via architecture", func(t *testing.T) {
		rule, err := arch.StructsImplementInterfaces(".*Repository$", ".*RepositoryInterface$")
		if err != nil {
			t.Fatalf("Failed to create rule: %v", err)
		}
		if rule == nil {
			t.Fatal("Rule should not be nil")
		}
	})
}

// TestArchitecture_ValidateInterfaceImplementations tests validation of interface implementations
func TestArchitecture_ValidateInterfaceImplementations(t *testing.T) {
	arch, err := arctest.New("./testdata")
	if err != nil {
		t.Fatalf("Failed to create architecture: %v", err)
	}

	err = arch.ParsePackages()
	if err != nil {
		t.Fatalf("Failed to parse packages: %v", err)
	}

	t.Run("validation with violations", func(t *testing.T) {
		rule, err := arch.StructsImplementInterfaces(".*Repository$", ".*RepositoryInterface$")
		if err != nil {
			t.Fatalf("Failed to create rule: %v", err)
		}

		valid, violations := arch.ValidateInterfaceImplementations([]*arctest.InterfaceImplementationRule{rule})
		if valid {
			t.Fatal("Expected validation to fail due to ProductRepository not implementing ProductRepositoryInterface")
		}
		if len(violations) == 0 {
			t.Fatal("Expected violations, but found none")
		}
	})

	t.Run("validation without violations", func(t *testing.T) {
		// Only check UserRepository which properly implements its interface
		rule, err := arch.StructsImplementInterfaces("UserRepository$", "UserRepositoryInterface$")
		if err != nil {
			t.Fatalf("Failed to create rule: %v", err)
		}

		valid, violations := arch.ValidateInterfaceImplementations([]*arctest.InterfaceImplementationRule{rule})
		if !valid {
			t.Fatalf("Expected validation to pass, but found violations: %v", violations)
		}
	})
}

// TestArchitecture_FindAllImplementations tests finding all implementations of an interface
func TestArchitecture_FindAllImplementations(t *testing.T) {
	arch, err := arctest.New("./testdata")
	if err != nil {
		t.Fatalf("Failed to create architecture: %v", err)
	}

	err = arch.ParsePackages()
	if err != nil {
		t.Fatalf("Failed to parse packages: %v", err)
	}

	t.Run("find existing implementation", func(t *testing.T) {
		implementations, err := arch.FindAllImplementations("UserRepositoryInterface", "domain")
		if err != nil {
			t.Fatalf("Failed to find implementations: %v", err)
		}

		if len(implementations) == 0 {
			t.Fatal("Expected to find implementations of UserRepositoryInterface, but found none")
		}

		foundUserRepo := false
		for _, impl := range implementations {
			if impl.Name == "UserRepository" {
				foundUserRepo = true
				break
			}
		}

		if !foundUserRepo {
			t.Fatal("Expected to find UserRepository as implementation of UserRepositoryInterface")
		}
	})

	t.Run("with non-existent interface", func(t *testing.T) {
		_, err := arch.FindAllImplementations("NonExistentInterface", "domain")
		if err == nil {
			t.Fatal("Expected error for non-existent interface, got nil")
		}
	})

	t.Run("with non-existent package", func(t *testing.T) {
		_, err := arch.FindAllImplementations("UserRepositoryInterface", "non_existent")
		if err == nil {
			t.Fatal("Expected error for non-existent package, got nil")
		}
	})
}
