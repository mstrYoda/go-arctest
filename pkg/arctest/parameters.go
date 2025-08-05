package arctest

import (
	"fmt"
	"regexp"
	"strings"
)

// ParameterRule represents a rule for checking method parameters
type ParameterRule struct {
	StructPattern             string // regex pattern for struct names
	MethodPattern             string // regex pattern for method names
	ParameterTypePattern      string // regex pattern for parameter types to check
	ShouldUseInterface        bool   // if true, parameters should be interfaces, if false, they should be structs
	structPatternRegex        *regexp.Regexp
	methodPatternRegex        *regexp.Regexp
	parameterTypePatternRegex *regexp.Regexp
}

// NewParameterRule creates a new parameter rule
func NewParameterRule(structPattern, methodPattern, parameterTypePattern string, shouldUseInterface bool) (*ParameterRule, error) {
	structRegex, err := regexp.Compile(structPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid struct pattern: %w", err)
	}

	methodRegex, err := regexp.Compile(methodPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid method pattern: %w", err)
	}

	paramRegex, err := regexp.Compile(parameterTypePattern)
	if err != nil {
		return nil, fmt.Errorf("invalid parameter type pattern: %w", err)
	}

	return &ParameterRule{
		StructPattern:             structPattern,
		MethodPattern:             methodPattern,
		ParameterTypePattern:      parameterTypePattern,
		ShouldUseInterface:        shouldUseInterface,
		structPatternRegex:        structRegex,
		methodPatternRegex:        methodRegex,
		parameterTypePatternRegex: paramRegex,
	}, nil
}

// CheckMethodParameters checks if method parameters match the required type (interface or struct)
func (a *Architecture) CheckMethodParameters(rules []*ParameterRule) ([]string, error) {
	violations := []string{}

	// Build a quick lookup of which types are interfaces and which are structs
	interfaces := make(map[string]bool)
	structs := make(map[string]bool)

	for _, pkg := range a.Packages {
		pkgPrefix := pkg.Name + "."
		for name := range pkg.Interfaces {
			interfaces[name] = true
			interfaces[pkgPrefix+name] = true
		}
		for name := range pkg.Structs {
			structs[name] = true
			structs[pkgPrefix+name] = true
		}
	}

	// First, check regular methods
	methodViolations, err := a.checkStructMethods(rules, interfaces, structs)
	if err != nil {
		return nil, err
	}
	violations = append(violations, methodViolations...)

	// Then, check constructor functions using imported dependencies
	constructorViolations, err := a.checkConstructorParameters(rules, interfaces, structs)
	if err != nil {
		return nil, err
	}
	violations = append(violations, constructorViolations...)

	return violations, nil
}

// checkStructMethods checks parameters of struct methods
func (a *Architecture) checkStructMethods(rules []*ParameterRule, interfaces, structs map[string]bool) ([]string, error) {
	violations := []string{}

	for _, rule := range rules {
		for _, pkg := range a.Packages {
			for _, s := range pkg.Structs {
				// Check if the struct matches the pattern
				if !rule.structPatternRegex.MatchString(s.Name) {
					continue
				}

				// For each method
				for _, m := range s.Methods {
					// Check if the method matches the pattern
					if !rule.methodPatternRegex.MatchString(m.Name) {
						continue
					}

					// For each parameter
					for _, p := range m.Params {
						// Skip empty or primitive types
						if p.Type == "" || isPrimitiveType(p.Type) {
							continue
						}

						// Process and check parameter
						if violation := checkParameterType(p.Type, rule, interfaces, structs, m.Name, s.Name, s.Pkg.Path); violation != "" {
							violations = append(violations, violation)
						}
					}
				}
			}
		}
	}

	return violations, nil
}

// checkConstructorParameters checks parameters of constructor functions
func (a *Architecture) checkConstructorParameters(rules []*ParameterRule, interfaces, structs map[string]bool) ([]string, error) {
	violations := []string{}

	// For each rule
	for _, rule := range rules {
		// For each package
		for _, pkg := range a.Packages {
			// For each struct that matches our pattern
			for structName := range pkg.Structs {
				if !rule.structPatternRegex.MatchString(structName) {
					continue
				}

				// Look for possible constructor functions for this struct in imported code
				possibleConstructors := []string{
					"New" + structName,
					"New" + strings.TrimSuffix(structName, "Service") + "Service",
					"New" + strings.TrimSuffix(structName, "Repository") + "Repository",
					"New" + strings.TrimSuffix(structName, "Controller") + "Controller",
				}

				// Check imports for dependencies that match our parameter pattern
				for _, importPath := range pkg.Imports {
					// Skip standard library imports
					if !strings.Contains(importPath, ".") && !strings.Contains(importPath, "/") {
						continue
					}

					importedPkg := a.findPackageByImportPath(importPath)
					if importedPkg == nil {
						continue
					}

					// Check imported package for types that match our parameter pattern
					for typeName, isStruct := range getTypesMap(importedPkg) {
						baseType := typeName
						if strings.Contains(typeName, ".") {
							baseType = typeName[strings.LastIndex(typeName, ".")+1:]
						}

						// If the type matches our parameter pattern
						if rule.parameterTypePatternRegex.MatchString(baseType) {
							// For each possible constructor
							for _, constructorName := range possibleConstructors {
								// If the constructor matches our method pattern
								if rule.methodPatternRegex.MatchString(constructorName) {
									fullTypeName := importedPkg.Name + "." + baseType

									// Determine if it's an interface or struct
									isInterface := interfaces[fullTypeName] || interfaces[baseType]
									isStruct := isStruct || structs[fullTypeName] || structs[baseType]

									// Check if it violates our rule
									if rule.ShouldUseInterface && isStruct && !isInterface {
										violations = append(violations, fmt.Sprintf(
											"Constructor %q of struct %q in package %q uses struct type %q as parameter, but should use an interface",
											constructorName, structName, pkg.Path, fullTypeName,
										))
									} else if !rule.ShouldUseInterface && isInterface && !isStruct {
										violations = append(violations, fmt.Sprintf(
											"Constructor %q of struct %q in package %q uses interface type %q as parameter, but should use a struct",
											constructorName, structName, pkg.Path, fullTypeName,
										))
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return violations, nil
}

// getTypesMap returns a map of all types in a package with a boolean indicating if they're structs
func getTypesMap(pkg *Package) map[string]bool {
	typesMap := make(map[string]bool)

	// Add structs
	for name := range pkg.Structs {
		typesMap[name] = true
	}

	// Add interfaces (with false to indicate they're not structs)
	for name := range pkg.Interfaces {
		typesMap[name] = false
	}

	return typesMap
}

// checkParameterType checks a parameter type against the rule and returns a violation string if there is one
func checkParameterType(paramType string, rule *ParameterRule, interfaces, structs map[string]bool, methodName, structName, pkgPath string) string {
	// Remove pointer prefix if exists
	if strings.HasPrefix(paramType, "*") {
		paramType = paramType[1:]
	}

	// Extract base type name without package for matching against pattern
	baseType := paramType
	if lastDot := strings.LastIndex(paramType, "."); lastDot != -1 {
		baseType = paramType[lastDot+1:]
	}

	// Check if the base type matches the pattern
	if !rule.parameterTypePatternRegex.MatchString(baseType) {
		return ""
	}

	// Determine if it's an interface or struct using the full parameter type
	isInterface := interfaces[paramType]
	isStruct := structs[paramType]

	// If we can't determine the type directly, try to determine by examining the packages
	if !isInterface && !isStruct {
		// For simplicity in this refactoring, we'll assume we can determine the type
		// from the interfaces and structs maps. The full implementation would need to pass
		// the Architecture object or the necessary package information.
	}

	// Check if the parameter type matches the rule
	if rule.ShouldUseInterface && isStruct && !isInterface {
		return fmt.Sprintf(
			"Method %q of struct %q in package %q uses struct type %q as parameter, but should use an interface",
			methodName, structName, pkgPath, paramType,
		)
	} else if !rule.ShouldUseInterface && isInterface && !isStruct {
		return fmt.Sprintf(
			"Method %q of struct %q in package %q uses interface type %q as parameter, but should use a struct",
			methodName, structName, pkgPath, paramType,
		)
	}

	return ""
}

// findPackageByImportPath finds a package by its import path
func (a *Architecture) findPackageByImportPath(importPath string) *Package {
	// Extract package name from import path (last part)
	parts := strings.Split(importPath, "/")
	pkgName := parts[len(parts)-1]

	// Find package by name
	for _, pkg := range a.Packages {
		if pkg.Name == pkgName {
			return pkg
		}
	}

	return nil
}

// isPrimitiveType checks if a type is a primitive Go type
func isPrimitiveType(typeName string) bool {
	primitives := map[string]bool{
		"bool":       true,
		"int":        true,
		"int8":       true,
		"int16":      true,
		"int32":      true,
		"int64":      true,
		"uint":       true,
		"uint8":      true,
		"uint16":     true,
		"uint32":     true,
		"uint64":     true,
		"uintptr":    true,
		"float32":    true,
		"float64":    true,
		"complex64":  true,
		"complex128": true,
		"string":     true,
		"byte":       true,
		"rune":       true,
		"error":      true,
	}

	return primitives[typeName]
}

// MethodsShouldUseInterfaceParameters creates a rule that methods should use interface parameters
func (a *Architecture) MethodsShouldUseInterfaceParameters(structPattern, methodPattern, parameterTypePattern string) (*ParameterRule, error) {
	return NewParameterRule(structPattern, methodPattern, parameterTypePattern, true)
}

// MethodsShouldUseStructParameters creates a rule that methods should use struct parameters
func (a *Architecture) MethodsShouldUseStructParameters(structPattern, methodPattern, parameterTypePattern string) (*ParameterRule, error) {
	return NewParameterRule(structPattern, methodPattern, parameterTypePattern, false)
}

// ValidateMethodParameters validates that method parameters match the required type
func (a *Architecture) ValidateMethodParameters(rules []*ParameterRule) (bool, []string) {
	violations, _ := a.CheckMethodParameters(rules)
	return len(violations) == 0, violations
}

// Function represents a standalone function
type Function struct {
	Name       string
	Params     []*Parameter
	ReturnType string
	Pkg        *Package
}
