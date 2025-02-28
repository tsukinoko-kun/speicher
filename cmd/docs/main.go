package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"sort"
)

// nodeToString pretty–prints an AST node using the provided file set.
func nodeToString(fset *token.FileSet, node interface{}) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		return ""
	}
	return buf.String()
}

// FunctionInfo holds details about a function/method.
type FunctionInfo struct {
	Name      string
	Doc       string
	Signature string
}

// extractInterfaceMethods returns the list of methods declared
// directly on an interface type.
func extractInterfaceMethods(fset *token.FileSet, typ *doc.Type) []FunctionInfo {
	methods := []FunctionInfo{}
	// typ.Decl is a *ast.GenDecl. We expect one TypeSpec.
	if typ.Decl == nil || len(typ.Decl.Specs) == 0 {
		return methods
	}
	ts, ok := typ.Decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return methods
	}
	iface, ok := ts.Type.(*ast.InterfaceType)
	if !ok {
		return methods
	}
	// Iterate over the fields in the interface.
	for _, field := range iface.Methods.List {
		// Embedded interfaces have no names; skip them.
		if len(field.Names) == 0 {
			continue
		}
		for _, name := range field.Names {
			// Only include exported interface methods.
			if !ast.IsExported(name.Name) {
				continue
			}
			methodDoc := ""
			if field.Doc != nil {
				methodDoc = field.Doc.Text()
			} else if field.Comment != nil {
				methodDoc = field.Comment.Text()
			}
			methodSig := interfaceMethodSignature(fset, field, name.Name)
			methods = append(methods, FunctionInfo{
				Name:      name.Name,
				Doc:       methodDoc,
				Signature: methodSig,
			})
		}
	}
	return methods
}

func interfaceMethodSignature(
	fset *token.FileSet, field *ast.Field, methodName string,
) string {
	if ftype, ok := field.Type.(*ast.FuncType); ok {
		// Create a fake function declaration for formatting.
		fn := &ast.FuncDecl{
			Name: ast.NewIdent(methodName),
			Type: ftype,
		}
		return nodeToString(fset, fn)
	}
	return nodeToString(fset, field)
}

func main() {
	fset := token.NewFileSet()
	// Parse the current directory with comments.
	pkgs, err := parser.ParseDir(fset, ".", nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Error parsing directory: %v", err)
	}

	// Use the first package found.
	var apkg *doc.Package
	for _, pkg := range pkgs {
		// The "./" import path is just a placeholder.
		apkg = doc.New(pkg, "./", doc.AllDecls)
		break
	}
	if apkg == nil {
		log.Fatal("No package found")
	}

	// Package header.
	fmt.Printf("# Package %s\n\n", apkg.Name)
	if apkg.Doc != "" {
		fmt.Printf("%s\n\n", apkg.Doc)
	}

	// Document package-level functions.
	if len(apkg.Funcs) > 0 {
		fmt.Print("## Functions\n\n")
		sort.Slice(apkg.Funcs, func(i, j int) bool {
			return apkg.Funcs[i].Name < apkg.Funcs[j].Name
		})
		for _, fn := range apkg.Funcs {
			fmt.Printf("### %s\n\n", fn.Name)
			signature := nodeToString(fset, fn.Decl)
			if fn.Doc != "" {
				fmt.Printf("%s\n\n", fn.Doc)
			}
			fmt.Printf("```go\n%s\n```\n\n", signature)
		}
	}

	// Document only exported types.
	if len(apkg.Types) > 0 {
		fmt.Print("## Types\n\n")
		// Filter out private types.
		var types []*doc.Type
		for _, typ := range apkg.Types {
			if !ast.IsExported(typ.Name) {
				continue
			}
			types = append(types, typ)
		}
		sort.Slice(types, func(i, j int) bool {
			return types[i].Name < types[j].Name
		})
		for _, typ := range types {
			fmt.Printf("### %s\n\n", typ.Name)
			signature := nodeToString(fset, typ.Decl)
			if typ.Doc != "" {
				fmt.Printf("%s\n\n", typ.Doc)
			}
			if len(signature) < 100 {
				fmt.Printf("```go\n%s\n```\n\n", signature)
			}

			// Gather methods.
			methodMap := make(map[string]FunctionInfo)
			// For interface types, extract methods declared in the type literal.
			var ts *ast.TypeSpec
			if typ.Decl != nil && len(typ.Decl.Specs) > 0 {
				ts, _ = typ.Decl.Specs[0].(*ast.TypeSpec)
			}
			if ts != nil {
				if _, ok := ts.Type.(*ast.InterfaceType); ok {
					ifaceMethods := extractInterfaceMethods(fset, typ)
					for _, m := range ifaceMethods {
						methodMap[m.Name] = m
					}
				}
			}
			// Include methods associated via separate functions.
			for _, m := range typ.Methods {
				// Even though these are methods on a type, check again if they are
				// exported.
				if !ast.IsExported(m.Name) {
					continue
				}
				mSig := nodeToString(fset, m.Decl)
				methodMap[m.Name] = FunctionInfo{
					Name:      m.Name,
					Doc:       m.Doc,
					Signature: mSig,
				}
			}

			// If there are methods, document them.
			if len(methodMap) > 0 {
				fmt.Print("#### Methods\n\n")
				// Sort methods alphabetically.
				var mNames []string
				for name := range methodMap {
					mNames = append(mNames, name)
				}
				sort.Strings(mNames)
				for _, name := range mNames {
					m := methodMap[name]
					fmt.Printf("##### %s\n\n", m.Name)
					if m.Doc != "" {
						fmt.Printf("%s\n\n", m.Doc)
					}
					if m.Signature != "" {
						fmt.Printf("```go\n%s\n```\n\n", m.Signature)
					}
				}
			}
		}
	}
}
