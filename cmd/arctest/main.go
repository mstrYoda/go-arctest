package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mstrYoda/go-arctest/pkg/config"
)

func main() {
	// Define command line flags
	configPath := flag.String("config", ".arctest.yml", "Path to the configuration file")
	projectPath := flag.String("project", ".", "Path to the project root")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	// Resolve absolute paths
	absProjectPath, err := filepath.Abs(*projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving project path: %v\n", err)
		os.Exit(1)
	}

	absConfigPath := *configPath
	if !filepath.IsAbs(absConfigPath) {
		absConfigPath = filepath.Join(absProjectPath, absConfigPath)
	}

	// Load the configuration
	cfg, err := config.LoadConfig(absConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Run the architecture tests
	passed, violations, err := cfg.RunArchitectureTests(absProjectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running architecture tests: %v\n", err)
		os.Exit(1)
	}

	// Print the results
	if passed {
		fmt.Println("✅ Architecture tests passed!")
		os.Exit(0)
	} else {
		fmt.Println("❌ Architecture tests failed!")
		fmt.Println("\nViolations:")
		for _, violation := range violations {
			fmt.Printf("  - %s\n", violation)
		}

		if *verbose {
			fmt.Println("\nConfiguration:")
			fmt.Println("Layers:")
			for _, layer := range cfg.Layers {
				fmt.Printf("  - %s (pattern: %s)\n", layer.Name, layer.Pattern)
			}

			fmt.Println("\nLayer Dependency Rules:")
			for _, rule := range cfg.Rules {
				fmt.Printf("  - %s -> %s\n", rule.From, rule.To)
			}

			if len(cfg.InterfaceRules) > 0 {
				fmt.Println("\nInterface Implementation Rules:")
				for _, rule := range cfg.InterfaceRules {
					fmt.Printf("  - Structs matching '%s' must implement interfaces matching '%s'\n",
						rule.StructPattern, rule.InterfacePattern)
				}
			}

			if len(cfg.ParameterRules) > 0 {
				fmt.Println("\nParameter Type Rules:")
				for _, rule := range cfg.ParameterRules {
					paramType := "struct"
					if rule.ShouldUseInterface {
						paramType = "interface"
					}
					fmt.Printf("  - Methods in '%s' matching '%s' should use %s parameters for '%s'\n",
						rule.StructPattern, rule.MethodPattern, paramType, rule.ParameterTypePattern)
				}
			}

			if len(cfg.LayerSpecificRules) > 0 {
				fmt.Println("\nLayer-Specific Rules:")
				for i, rule := range cfg.LayerSpecificRules {
					fmt.Printf("  - [%d] Layer '%s' rule type '%s'\n", i+1, rule.Layer, rule.RuleType)
					for k, v := range rule.Parameters {
						fmt.Printf("      %s: %s\n", k, v)
					}
				}
			}

			if len(cfg.DirectLayerDependencyRules) > 0 {
				fmt.Println("\nDirect Layer Dependency Rules:")
				for _, rule := range cfg.DirectLayerDependencyRules {
					action := "must not depend on"
					if rule.Allowed {
						action = "may depend on"
					}
					fmt.Printf("  - Layer '%s' %s layer '%s'\n", rule.SourceLayer, action, rule.TargetLayer)
				}
			}
		}

		os.Exit(1)
	}
}
