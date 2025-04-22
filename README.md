# Go ArcTest

Go ArcTest is a library for testing the architecture of your Go projects, similar to Java's ArchUnit. It provides tooling to validate architectural rules and constraints in your codebase through tests.

## Features

- **Package Dependency Analysis**: Check if a layer/package imports/depends on another package. Define rules for allowed and disallowed dependencies.
- **Interface Implementation Validation**: Ensure that specific structs implement required interfaces.
- **Parameter Type Checking**: Verify that method parameters use interfaces instead of concrete struct implementations, promoting loose coupling.
- **Layered Architecture Support**: Define layers and rules between them to enforce a clean layered architecture.
- **Layer-Specific Rules**: Define architectural rules specific to individual layers.
- **Direct Layer Dependency Rules**: Specify that one layer should not depend on another layer using a more intuitive API.
- **Nested Package Support**: Automatically analyzes nested packages within layers and applies architecture rules to them.

## Installation

```bash
go get github.com/mstrYoda/go-arctest
```

## Usage

### Basic Architecture Analysis

```go
// Initialize architecture with project root
arch, err := arctest.New("./")
if err != nil {
    t.Fatalf("Failed to create architecture: %v", err)
}

// Parse all packages in the project
err = arch.ParsePackages()
if err != nil {
    t.Fatalf("Failed to parse packages: %v", err)
}

// Parse specific packages and their subpackages
// err = arch.ParsePackages("internal/domain", "internal/service")
```

### Checking Package Dependencies

```go
// Create a rule that one package should not depend on another
rule, err := arch.DoesNotDependOn("^domain/.*$", "^infrastructure/.*$")
if err != nil {
    t.Fatalf("Failed to create dependency rule: %v", err)
}

// Validate dependencies
valid, violations := arch.ValidateDependenciesWithRules([]*arctest.DependencyRule{rule})
if !valid {
    for _, violation := range violations {
        t.Errorf("Dependency violation: %s", violation)
    }
}
```

### Checking Interface Implementations

```go
// Create a rule that all repository structs must implement repository interfaces
rule, err := arch.StructsImplementInterfaces(".*Repository$", ".*RepositoryInterface$")
if err != nil {
    t.Fatalf("Failed to create interface implementation rule: %v", err)
}

// Validate interface implementations
valid, violations := arch.ValidateInterfaceImplementations([]*arctest.InterfaceImplementationRule{rule})
if !valid {
    for _, violation := range violations {
        t.Errorf("Interface implementation violation: %s", violation)
    }
}
```

### Checking Method Parameters

```go
// Create a rule that all service methods should use interfaces as parameters
rule, err := arch.MethodsShouldUseInterfaceParameters(".*Service$", ".*", ".*Repository$")
if err != nil {
    t.Fatalf("Failed to create parameter rule: %v", err)
}

// Validate parameter types
valid, violations := arch.ValidateMethodParameters([]*arctest.ParameterRule{rule})
if !valid {
    for _, violation := range violations {
        t.Errorf("Parameter type violation: %s", violation)
    }
}
```

### Testing for Interface Parameter Violations

The following example demonstrates how to create a test that verifies methods are using interfaces as parameters instead of concrete struct implementations. This is useful for enforcing the Dependency Inversion Principle.

```go
func TestInterfaceParameterViolation(t *testing.T) {
    // Initialize architecture and parse packages
    arch, _ := arctest.New("./example_project")
    arch.ParsePackages("domain", "utils")
    
    // Define the rule: methods named "Update*" should use interface parameters for "Logger"
    // This will check if methods are violating the rule by using concrete structs
    rule, _ := arch.MethodsShouldUseInterfaceParameters(".*Service.*", "Update.*", ".*Logger")
    
    // Validate parameter types - we expect violations
    valid, violations := arch.ValidateMethodParameters([]*arctest.ParameterRule{rule})
    
    // If we're testing that our architecture rule is correctly detecting violations,
    // we would expect valid to be false and violations to be non-empty
    if valid {
        t.Error("Expected interface parameter violations, but none were found!")
    } else {
        t.Logf("Successfully detected methods using concrete types instead of interfaces:")
        for _, violation := range violations {
            t.Logf("  ✓ %s", violation)
        }
    }
    
    // You can also use the inverse rule to confirm that concrete types are being used
    structRule, _ := arch.MethodsShouldUseStructParameters(".*Service.*", "Update.*", ".*Logger")
    structValid, _ := arch.ValidateMethodParameters([]*arctest.ParameterRule{structRule})
    
    // This should pass as we expect to find methods using concrete structs
    if !structValid {
        t.Error("Methods are not using struct parameters as expected in our test case")
    }
}
```

### Defining and Checking a Layered Architecture

```go
// Define layers
domainLayer, _ := arctest.NewLayer("Domain", "^domain/.*$")
applicationLayer, _ := arctest.NewLayer("Application", "^application/.*$")
infrastructureLayer, _ := arctest.NewLayer("Infrastructure", "^infrastructure/.*$")
presentationLayer, _ := arctest.NewLayer("Presentation", "^presentation/.*$")

// Define layered architecture
layeredArch := arctest.NewLayeredArchitecture(
    domainLayer,
    applicationLayer,
    infrastructureLayer,
    presentationLayer,
)

// Define dependency rules
layeredArch.AddRule("Application", "Domain")
layeredArch.AddRule("Infrastructure", "Domain")
layeredArch.AddRule("Infrastructure", "Application")
layeredArch.AddRule("Presentation", "Domain")
layeredArch.AddRule("Presentation", "Application")
layeredArch.AddRule("Presentation", "Infrastructure")

// Check layered architecture
violations, err := layeredArch.Check(arch)
if err != nil {
    t.Fatalf("Failed to check layered architecture: %v", err)
}

for _, violation := range violations {
    t.Errorf("Architecture violation: %s", violation)
}
```

### Working with Nested Packages

The library automatically handles nested packages within layers. When you define a layer with a pattern like `^domain$`, it automatically includes subpackages like `domain/entities`, `domain/services`, etc.

```go
// Initialize architecture
arch, _ := arctest.New("./")

// This will recursively parse the packages and their subpackages
arch.ParsePackages("app", "domain")

// Define layers - these patterns will match the packages and their subpackages
appLayer, _ := arctest.NewLayer("App", "^app$") // Matches app, app/handlers, app/services, etc.
domainLayer, _ := arctest.NewLayer("Domain", "^domain$") // Matches domain and all subpackages

// Create a layered architecture
layeredArch := arctest.NewLayeredArchitecture(appLayer, domainLayer)
layeredArch.SetArchitecture(arch)

// Define allowed dependencies
appLayer.DependsOnLayer(domainLayer, layeredArch)

// Run the check - this will check all packages and subpackages
violations, _ := layeredArch.Check(arch)

// The result will include violations from all subpackages
for _, violation := range violations {
    t.Errorf("Architecture violation: %s", violation)
}
```

### Using Layer-Specific Rules

You can define architectural rules specific to individual layers, which is often more intuitive and clearer than defining rules at the architecture level.

```go
// Set architecture for the layered architecture
layeredArch.SetArchitecture(arch)

// Define layer dependencies using layer-specific methods
applicationLayer.DependsOn("Domain", layeredArch)
infrastructureLayer.DependsOn("Domain", layeredArch)
presentationLayer.DependsOn("Domain", layeredArch)
presentationLayer.DependsOn("Application", layeredArch)

// Layer-specific rules:

// 1. Domain layer should not depend on any other layer
rule1, err := domainLayer.DoesNotDependOn("^(application|infrastructure|presentation)$")

// 2. Infrastructure repositories should implement domain interfaces
rule2, err := infrastructureLayer.StructsImplementInterfaces(".*Repository$", ".*RepositoryInterface$")

// 3. Application services should use interfaces as parameters
rule3, err := applicationLayer.MethodsShouldUseInterfaceParameters(".*Service$", "New.*", ".*Repository$")

// Run the tests with these layer-specific rules
valid1, violations1 := arch.ValidateDependenciesWithRules([]*arctest.DependencyRule{rule1})
valid2, violations2 := arch.ValidateInterfaceImplementations([]*arctest.InterfaceImplementationRule{rule2})
valid3, violations3 := arch.ValidateMethodParameters([]*arctest.ParameterRule{rule3})
```

### Using Direct Layer Dependency Rules

You can use a more intuitive API to define that one layer should not depend on another layer, without having to use regex patterns. This is particularly useful when working with a layered architecture.

```go
// Define layers
domainLayer, _ := arctest.NewLayer("Domain", "^domain$")
applicationLayer, _ := arctest.NewLayer("Application", "^application$")
utilsLayer, _ := arctest.NewLayer("Utils", "^utils$")

// Set architecture for the layers
layeredArch := arctest.NewLayeredArchitecture(domainLayer, applicationLayer, utilsLayer)
layeredArch.SetArchitecture(arch)

// Method 1: Define direct layer dependencies
// Domain should not depend on Application layer
domainAppRule, err := domainLayer.DoesNotDependOnLayer(applicationLayer)
if err != nil {
    t.Fatalf("Failed to create layer dependency rule: %v", err)
}

// Domain should not depend on Utils layer
domainUtilsRule, err := domainLayer.DoesNotDependOnLayer(utilsLayer)
if err != nil {
    t.Fatalf("Failed to create layer dependency rule: %v", err)
}

// Test the rules
valid, violations := arch.ValidateDependenciesWithRules([]*arctest.DependencyRule{
    domainAppRule,
    domainUtilsRule,
})

// Method 2: Define allowed dependencies between layers
// Application layer can depend on Domain layer
err = applicationLayer.DependsOnLayer(domainLayer, layeredArch)
if err != nil {
    t.Fatalf("Failed to create layer dependency: %v", err)
}

// Check layered architecture for violations
violations, err := layeredArch.Check(arch)
```

### Testing for Dependency Violations

The following example demonstrates how to create a test that checks for dependency violations. This is useful for TDD (Test-Driven Development) of your architecture, where you might want to verify that a dependency rule is properly enforced.

```go
func TestDependencyViolation(t *testing.T) {
    // Initialize architecture and parse packages
    arch, _ := arctest.New("./example_project")
    arch.ParsePackages("domain", "utils") // domain should not depend on utils
    
    // Define layers
    domainLayer, _ := arctest.NewLayer("Domain", "^domain$")
    utilsLayer, _ := arctest.NewLayer("Utils", "^utils$")
    
    // Create a rule that domain should not depend on utils
    layeredArch := arctest.NewLayeredArchitecture(domainLayer, utilsLayer)
    layeredArch.SetArchitecture(arch)
    
    // Create the rule using the layer-specific method
    rule, _ := domainLayer.DoesNotDependOn("^utils$")
    
    // Validate dependencies - we expect violations if domain imports utils
    valid, violations := arch.ValidateDependenciesWithRules([]*arctest.DependencyRule{rule})
    
    // For a test that expects violations (like testing that our rules work)
    // we'd check that valid is false and violations contains expected issues
    if valid {
        t.Error("Expected dependency violations, but none were found!")
    } else {
        t.Logf("Successfully detected dependency violations:")
        for _, violation := range violations {
            t.Logf("  ✓ %s", violation)
        }
    }
}
```

## Example

See the `examples` directory for a complete example of how to use this library in your architecture tests.

## License

MIT 