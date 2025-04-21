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

	// For each rule
	for _, rule := range rules {
		// For each package
		for _, pkg := range a.Packages {
			// For each struct
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

						// Remove pointer prefix if exists
						paramType := p.Type
						if strings.HasPrefix(paramType, "*") {
							paramType = paramType[1:]
						}

						// Check if the parameter type matches the pattern
						if !rule.parameterTypePatternRegex.MatchString(paramType) {
							continue
						}

						isInterface := interfaces[paramType]
						isStruct := structs[paramType]

						// If we can't determine the type, skip it
						if !isInterface && !isStruct {
							continue
						}

						// Check if the parameter type matches the rule
						if rule.ShouldUseInterface && !isInterface {
							violations = append(violations, fmt.Sprintf(
								"Method %q of struct %q in package %q uses struct type %q as parameter, but should use an interface",
								m.Name, s.Name, s.Pkg.Path, paramType,
							))
						} else if !rule.ShouldUseInterface && !isStruct {
							violations = append(violations, fmt.Sprintf(
								"Method %q of struct %q in package %q uses interface type %q as parameter, but should use a struct",
								m.Name, s.Name, s.Pkg.Path, paramType,
							))
						}
					}
				}
			}
		}
	}

	return violations, nil
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
