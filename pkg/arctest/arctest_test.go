package arctest_test

import (
	"path/filepath"
	"testing"

	"github.com/mstrYoda/go-arctest/pkg/arctest"
)

// TestNew tests creation of Architecture instances
func TestNew(t *testing.T) {
	t.Run("successfully create architecture", func(t *testing.T) {
		arch, err := arctest.New("./testdata")
		if err != nil {
			t.Fatalf("Failed to create architecture: %v", err)
		}
		if arch == nil {
			t.Fatal("Architecture should not be nil")
		}
	})

	t.Run("with invalid path", func(t *testing.T) {
		// Create a path that definitely doesn't exist
		nonExistentPath := filepath.Join(t.TempDir(), "this_does_not_exist")
		// Try to use the non-existent path
		_, err := arctest.New(nonExistentPath)
		if err == nil {
			t.Fatal("Expected error for non-existent path, got nil")
		}
	})
}

// TestArchitecture_ParsePackages tests parsing packages
func TestArchitecture_ParsePackages(t *testing.T) {
	arch, err := arctest.New("./testdata")
	if err != nil {
		t.Fatalf("Failed to create architecture: %v", err)
	}

	t.Run("parse specific packages", func(t *testing.T) {
		err := arch.ParsePackages("domain", "application")
		if err != nil {
			t.Fatalf("Failed to parse packages: %v", err)
		}

		// Verify the packages were parsed
		pkg := arch.GetPackage("domain")
		if pkg == nil {
			t.Fatal("Expected domain package to be parsed")
		}

		// Check for expected types in domain package
		if len(pkg.Structs) == 0 {
			t.Fatal("Expected structs in domain package")
		}

		if _, ok := pkg.Structs["UserEntity"]; !ok {
			t.Fatal("Expected UserEntity struct in domain package")
		}

		if len(pkg.Interfaces) == 0 {
			t.Fatal("Expected interfaces in domain package")
		}

		if _, ok := pkg.Interfaces["UserRepositoryInterface"]; !ok {
			t.Fatal("Expected UserRepositoryInterface in domain package")
		}
	})

	t.Run("parse all packages", func(t *testing.T) {
		// Create new architecture to avoid interference with previous tests
		arch, err := arctest.New("./testdata")
		if err != nil {
			t.Fatalf("Failed to create architecture: %v", err)
		}

		err = arch.ParsePackages()
		if err != nil {
			t.Fatalf("Failed to parse all packages: %v", err)
		}

		// Check that all packages were parsed
		packages := []string{"domain", "application", "infrastructure", "presentation"}
		for _, pkgName := range packages {
			pkg := arch.GetPackage(pkgName)
			if pkg == nil {
				t.Fatalf("Expected %s package to be parsed", pkgName)
			}
		}
	})

	t.Run("with invalid package", func(t *testing.T) {
		err := arch.ParsePackages("non_existent_package")
		if err == nil {
			t.Fatal("Expected error for non-existent package, got nil")
		}
	})
}

// TestArchitecture_ParsePackage tests parsing a single package
func TestArchitecture_ParsePackage(t *testing.T) {
	arch, err := arctest.New("./testdata")
	if err != nil {
		t.Fatalf("Failed to create architecture: %v", err)
	}

	t.Run("parse infrastructure package", func(t *testing.T) {
		err := arch.ParsePackage("infrastructure")
		if err != nil {
			t.Fatalf("Failed to parse infrastructure package: %v", err)
		}

		pkg := arch.GetPackage("infrastructure")
		if pkg == nil {
			t.Fatal("Expected infrastructure package to be parsed")
		}

		// Verify the package contains the expected structs
		if _, ok := pkg.Structs["UserRepository"]; !ok {
			t.Fatal("Expected UserRepository struct in infrastructure package")
		}

		// Verify imports
		foundDomainImport := false
		for _, imp := range pkg.Imports {
			if imp == "github.com/mstrYoda/go-arctest/pkg/arctest/testdata/domain" {
				foundDomainImport = true
				break
			}
		}
		if !foundDomainImport {
			t.Fatal("Expected infrastructure package to import domain")
		}
	})

	t.Run("with non-go file", func(t *testing.T) {
		// Create a temporary file in testdata that's not a Go file
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "not_a_go_file.txt")
		err := arch.ParsePackage(tempFile)
		if err == nil {
			t.Fatal("Expected error for non-Go file, got nil")
		}
	})
}

// TestArchitecture_GetPackage tests retrieving a package
func TestArchitecture_GetPackage(t *testing.T) {
	arch, err := arctest.New("./testdata")
	if err != nil {
		t.Fatalf("Failed to create architecture: %v", err)
	}

	err = arch.ParsePackages("domain")
	if err != nil {
		t.Fatalf("Failed to parse domain package: %v", err)
	}

	t.Run("get existing package", func(t *testing.T) {
		pkg := arch.GetPackage("domain")
		if pkg == nil {
			t.Fatal("Expected domain package, got nil")
		}
		if pkg.Name != "domain" {
			t.Fatalf("Expected package name 'domain', got '%s'", pkg.Name)
		}
	})

	t.Run("get non-existent package", func(t *testing.T) {
		pkg := arch.GetPackage("non_existent")
		if pkg != nil {
			t.Fatalf("Expected nil for non-existent package, got %v", pkg)
		}
	})
}
