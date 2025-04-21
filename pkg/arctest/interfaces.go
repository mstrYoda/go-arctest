package arctest

import (
	"fmt"
	"regexp"
)

// InterfaceImplementationRule represents a rule that structs must implement interfaces
type InterfaceImplementationRule struct {
	StructPattern         string // regex pattern for struct names
	InterfacePattern      string // regex pattern for interface names
	structPatternRegex    *regexp.Regexp
	interfacePatternRegex *regexp.Regexp
}

// NewInterfaceImplementationRule creates a new interface implementation rule
func NewInterfaceImplementationRule(structPattern, interfacePattern string) (*InterfaceImplementationRule, error) {
	structRegex, err := regexp.Compile(structPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid struct pattern: %w", err)
	}

	interfaceRegex, err := regexp.Compile(interfacePattern)
	if err != nil {
		return nil, fmt.Errorf("invalid interface pattern: %w", err)
	}

	return &InterfaceImplementationRule{
		StructPattern:         structPattern,
		InterfacePattern:      interfacePattern,
		structPatternRegex:    structRegex,
		interfacePatternRegex: interfaceRegex,
	}, nil
}

// CheckInterfaceImplementation checks if a struct implements an interface
func CheckInterfaceImplementation(s *Struct, i *Interface) bool {
	// If the interface has no methods, then any struct implements it
	if len(i.Methods) == 0 {
		return true
	}

	// Check if the struct has all the methods required by the interface
	for _, iMethod := range i.Methods {
		found := false
		for _, sMethod := range s.Methods {
			if sMethod.Name == iMethod.Name {
				// Check if the method signatures match
				if methodSignaturesMatch(sMethod, iMethod) {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// methodSignaturesMatch checks if two methods have matching signatures
// This is a simplified check that only ensures the method names match
// In a real implementation, you would also check parameter types and return types
func methodSignaturesMatch(m1, m2 *Method) bool {
	if m1.Name != m2.Name {
		return false
	}

	// For simplicity, we're just checking if the parameter count matches
	// In a real implementation, you would check parameter types as well
	if len(m1.Params) != len(m2.Params) {
		return false
	}

	// Check if both have return values or neither does
	if (m1.ReturnType == "") != (m2.ReturnType == "") {
		return false
	}

	return true
}

// CheckStructImplementsInterfaces checks all structs against the provided interface implementation rules
func (a *Architecture) CheckStructImplementsInterfaces(rules []*InterfaceImplementationRule) ([]string, error) {
	violations := []string{}

	// For each rule
	for _, rule := range rules {
		// Keep track of matching structs and interfaces
		matchingStructs := []*Struct{}
		matchingInterfaces := []*Interface{}

		// Find all structs and interfaces that match the pattern
		for _, pkg := range a.Packages {
			for _, s := range pkg.Structs {
				if rule.structPatternRegex.MatchString(s.Name) {
					matchingStructs = append(matchingStructs, s)
				}
			}

			for _, i := range pkg.Interfaces {
				if rule.interfacePatternRegex.MatchString(i.Name) {
					matchingInterfaces = append(matchingInterfaces, i)
				}
			}
		}

		// For each matching struct, check if it implements at least one matching interface
		for _, s := range matchingStructs {
			implementsAny := false
			for _, i := range matchingInterfaces {
				if CheckInterfaceImplementation(s, i) {
					implementsAny = true
					break
				}
			}

			if !implementsAny && len(matchingInterfaces) > 0 {
				violations = append(violations, fmt.Sprintf(
					"Struct %q in package %q does not implement any interface matching %q",
					s.Name, s.Pkg.Path, rule.InterfacePattern,
				))
			}
		}
	}

	return violations, nil
}

// StructsImplementInterfaces creates a rule that structs matching a pattern must implement interfaces matching a pattern
func (a *Architecture) StructsImplementInterfaces(structPattern, interfacePattern string) (*InterfaceImplementationRule, error) {
	return NewInterfaceImplementationRule(structPattern, interfacePattern)
}

// ValidateInterfaceImplementations validates that structs implement interfaces according to rules
func (a *Architecture) ValidateInterfaceImplementations(rules []*InterfaceImplementationRule) (bool, []string) {
	violations, _ := a.CheckStructImplementsInterfaces(rules)
	return len(violations) == 0, violations
}

// FindAllImplementations finds all structs that implement a given interface
func (a *Architecture) FindAllImplementations(interfaceName, interfacePkgPath string) ([]*Struct, error) {
	// Find the interface
	pkg := a.GetPackage(interfacePkgPath)
	if pkg == nil {
		return nil, fmt.Errorf("package %q not found", interfacePkgPath)
	}

	iface, found := pkg.Interfaces[interfaceName]
	if !found {
		return nil, fmt.Errorf("interface %q not found in package %q", interfaceName, interfacePkgPath)
	}

	// Find all structs that implement the interface
	implementations := []*Struct{}

	for _, pkg := range a.Packages {
		for _, s := range pkg.Structs {
			if CheckInterfaceImplementation(s, iface) {
				implementations = append(implementations, s)
			}
		}
	}

	return implementations, nil
}
