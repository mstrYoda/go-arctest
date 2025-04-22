package arctest

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Architecture represents a collection of packages and their relationships
type Architecture struct {
	Packages map[string]*Package
	basePath string
}

// Package represents a Go package with its imports and types
type Package struct {
	Name         string
	Path         string
	Imports      []string
	Structs      map[string]*Struct
	Interfaces   map[string]*Interface
	ImportedPkgs map[string]string // map of alias -> package path
}

// Struct represents a Go struct with its fields and methods
type Struct struct {
	Name    string
	Fields  []*Field
	Methods []*Method
	Pkg     *Package
}

// Field represents a struct field
type Field struct {
	Name string
	Type string
}

// Method represents a struct method
type Method struct {
	Name       string
	Params     []*Parameter
	ReturnType string
}

// Parameter represents a method parameter
type Parameter struct {
	Name string
	Type string
}

// Interface represents a Go interface with its methods
type Interface struct {
	Name    string
	Methods []*Method
	Pkg     *Package
}

// New creates a new Architecture instance for the given base path
func New(basePath string) (*Architecture, error) {
	abs, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return &Architecture{
		Packages: make(map[string]*Package),
		basePath: abs,
	}, nil
}

// ParsePackages parses all packages in the architecture
func (a *Architecture) ParsePackages(pkgPaths ...string) error {
	if len(pkgPaths) == 0 {
		// If no paths specified, parse all packages in the base path
		return a.parseAllPackages()
	}

	for _, path := range pkgPaths {
		if err := a.ParsePackage(path); err != nil {
			return err
		}
	}

	return nil
}

func (a *Architecture) parseAllPackages() error {
	return filepath.Walk(a.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && !strings.HasPrefix(info.Name(), ".") {
			relPath, err := filepath.Rel(a.basePath, path)
			if err != nil {
				return err
			}

			// Skip vendor directory and non-Go packages
			if relPath == "vendor" || strings.HasPrefix(relPath, "vendor/") {
				return filepath.SkipDir
			}

			// Check if directory contains .go files
			hasGoFiles := false
			files, err := os.ReadDir(path)
			if err != nil {
				return err
			}

			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") && !strings.HasSuffix(file.Name(), "_test.go") {
					hasGoFiles = true
					break
				}
			}

			if hasGoFiles {
				if err := a.ParsePackage(relPath); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// ParsePackage parses a specific package and its subpackages
func (a *Architecture) ParsePackage(pkgPath string) error {
	fullPath := filepath.Join(a.basePath, pkgPath)

	// First check if this is a directory
	info, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("failed to stat package path %s: %w", pkgPath, err)
	}

	if info.IsDir() {
		// Parse the current directory as a package
		if err := a.parsePackageDir(fullPath, pkgPath); err != nil {
			return err
		}

		// Now recursively parse all subdirectories that might contain Go packages
		files, err := os.ReadDir(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read package directory %s: %w", pkgPath, err)
		}

		for _, file := range files {
			if file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
				subPkgPath := filepath.Join(pkgPath, file.Name())
				// Check if the subdirectory contains any Go files before parsing
				hasGoFiles := false
				subDir := filepath.Join(fullPath, file.Name())
				subFiles, err := os.ReadDir(subDir)
				if err != nil {
					return fmt.Errorf("failed to read subdirectory %s: %w", subPkgPath, err)
				}

				for _, subFile := range subFiles {
					if !subFile.IsDir() && strings.HasSuffix(subFile.Name(), ".go") && !strings.HasSuffix(subFile.Name(), "_test.go") {
						hasGoFiles = true
						break
					}
				}

				if hasGoFiles {
					if err := a.ParsePackage(subPkgPath); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	// If it's not a directory, assume it's a Go file or pattern
	return a.parsePackageDir(filepath.Dir(fullPath), filepath.Dir(pkgPath))
}

// parsePackageDir parses a specific directory as a Go package
func (a *Architecture) parsePackageDir(fullPath, pkgPath string) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, fullPath, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ParseComments)

	if err != nil {
		return fmt.Errorf("failed to parse package %s: %w", pkgPath, err)
	}

	for pkgName, pkg := range pkgs {
		p := &Package{
			Name:         pkgName,
			Path:         pkgPath,
			Imports:      make([]string, 0),
			Structs:      make(map[string]*Struct),
			Interfaces:   make(map[string]*Interface),
			ImportedPkgs: make(map[string]string),
		}

		for _, file := range pkg.Files {
			// Process imports
			for _, imp := range file.Imports {
				importPath := strings.Trim(imp.Path.Value, "\"")
				fmt.Printf("Found import in %s: %s\n", pkgPath, importPath)
				p.Imports = append(p.Imports, importPath)

				// Handle import alias
				var alias string
				if imp.Name != nil {
					alias = imp.Name.Name
				} else {
					parts := strings.Split(importPath, "/")
					alias = parts[len(parts)-1]
				}
				p.ImportedPkgs[alias] = importPath
			}

			// Process declarations
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if ok && genDecl.Tok == token.TYPE {
					for _, spec := range genDecl.Specs {
						typeSpec, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}

						// Process struct types
						structType, isStruct := typeSpec.Type.(*ast.StructType)
						if isStruct {
							s := &Struct{
								Name:    typeSpec.Name.Name,
								Fields:  make([]*Field, 0),
								Methods: make([]*Method, 0),
								Pkg:     p,
							}

							// Process struct fields
							if structType.Fields != nil {
								for _, field := range structType.Fields.List {
									fieldType := ""
									// Get field type as string
									switch t := field.Type.(type) {
									case *ast.Ident:
										fieldType = t.Name
									case *ast.SelectorExpr:
										if x, ok := t.X.(*ast.Ident); ok {
											fieldType = x.Name + "." + t.Sel.Name
										}
									case *ast.StarExpr:
										// Handle pointer types
										switch pt := t.X.(type) {
										case *ast.Ident:
											fieldType = "*" + pt.Name
										case *ast.SelectorExpr:
											if x, ok := pt.X.(*ast.Ident); ok {
												fieldType = "*" + x.Name + "." + pt.Sel.Name
											}
										}
									}

									// Handle multiple names for the same type
									for _, name := range field.Names {
										s.Fields = append(s.Fields, &Field{
											Name: name.Name,
											Type: fieldType,
										})
									}
								}
							}

							p.Structs[s.Name] = s
						}

						// Process interface types
						interfaceType, isInterface := typeSpec.Type.(*ast.InterfaceType)
						if isInterface {
							i := &Interface{
								Name:    typeSpec.Name.Name,
								Methods: make([]*Method, 0),
								Pkg:     p,
							}

							// Process interface methods
							if interfaceType.Methods != nil {
								for _, method := range interfaceType.Methods.List {
									funcType, ok := method.Type.(*ast.FuncType)
									if !ok {
										continue
									}

									m := &Method{
										Name:       method.Names[0].Name,
										Params:     make([]*Parameter, 0),
										ReturnType: "",
									}

									// Process method parameters
									if funcType.Params != nil {
										for _, param := range funcType.Params.List {
											paramType := ""
											switch t := param.Type.(type) {
											case *ast.Ident:
												paramType = t.Name
											case *ast.SelectorExpr:
												if x, ok := t.X.(*ast.Ident); ok {
													paramType = x.Name + "." + t.Sel.Name
												}
											case *ast.StarExpr:
												// Handle pointer types
												switch pt := t.X.(type) {
												case *ast.Ident:
													paramType = "*" + pt.Name
												case *ast.SelectorExpr:
													if x, ok := pt.X.(*ast.Ident); ok {
														paramType = "*" + x.Name + "." + pt.Sel.Name
													}
												}
											}

											// Handle multiple names for the same type
											if len(param.Names) == 0 {
												m.Params = append(m.Params, &Parameter{
													Name: "",
													Type: paramType,
												})
											} else {
												for _, name := range param.Names {
													m.Params = append(m.Params, &Parameter{
														Name: name.Name,
														Type: paramType,
													})
												}
											}
										}
									}

									// Process return types
									if funcType.Results != nil && funcType.Results.List != nil {
										// For simplicity, just note if there's a return value
										m.ReturnType = "has_return"
									}

									i.Methods = append(i.Methods, m)
								}
							}

							p.Interfaces[i.Name] = i
						}
					}
				}
			}
		}

		// Find methods for structs
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				funcDecl, ok := decl.(*ast.FuncDecl)
				if !ok || funcDecl.Recv == nil {
					continue
				}

				// This is a method with a receiver
				recvType := ""
				if len(funcDecl.Recv.List) > 0 {
					switch rt := funcDecl.Recv.List[0].Type.(type) {
					case *ast.Ident:
						recvType = rt.Name
					case *ast.StarExpr:
						if ident, ok := rt.X.(*ast.Ident); ok {
							recvType = ident.Name
						}
					}
				}

				if recvType != "" {
					if s, found := p.Structs[recvType]; found {
						m := &Method{
							Name:       funcDecl.Name.Name,
							Params:     make([]*Parameter, 0),
							ReturnType: "",
						}

						// Process method parameters
						if funcDecl.Type.Params != nil {
							for _, param := range funcDecl.Type.Params.List {
								paramType := ""
								switch t := param.Type.(type) {
								case *ast.Ident:
									paramType = t.Name
								case *ast.SelectorExpr:
									if x, ok := t.X.(*ast.Ident); ok {
										paramType = x.Name + "." + t.Sel.Name
									}
								case *ast.StarExpr:
									// Handle pointer types
									switch pt := t.X.(type) {
									case *ast.Ident:
										paramType = "*" + pt.Name
									case *ast.SelectorExpr:
										if x, ok := pt.X.(*ast.Ident); ok {
											paramType = "*" + x.Name + "." + pt.Sel.Name
										}
									}
								}

								// Handle multiple names for the same type
								if len(param.Names) == 0 {
									m.Params = append(m.Params, &Parameter{
										Name: "",
										Type: paramType,
									})
								} else {
									for _, name := range param.Names {
										m.Params = append(m.Params, &Parameter{
											Name: name.Name,
											Type: paramType,
										})
									}
								}
							}
						}

						// Process return types
						if funcDecl.Type.Results != nil && funcDecl.Type.Results.List != nil {
							// For simplicity, just note if there's a return value
							m.ReturnType = "has_return"
						}

						s.Methods = append(s.Methods, m)
					}
				}
			}
		}

		a.Packages[pkgPath] = p
	}

	return nil
}

// GetPackage returns a package by path
func (a *Architecture) GetPackage(pkgPath string) *Package {
	return a.Packages[pkgPath]
}
