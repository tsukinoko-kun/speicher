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
	Name       string
	Doc        string
	Signature  string
	TypeParams string // Holds generic type parameter information (if any).
}

// interfaceMethodSignature formats an interface method as a fake function declaration.
// It also returns the type parameters (if any) as a separate string.
func interfaceMethodSignature(
	fset *token.FileSet, field *ast.Field, methodName string,
) (string, string) {
	var tp string
	if ftype, ok := field.Type.(*ast.FuncType); ok {
		// If there are type parameters, extract them.
		if ftype.TypeParams != nil {
			tp = nodeToString(fset, ftype.TypeParams)
		}
		// Create a fake function declaration for formatting.
		fn := &ast.FuncDecl{
			Name: ast.NewIdent(methodName),
			Type: ftype,
		}
		return nodeToString(fset, fn), tp
	}
	return nodeToString(fset, field), tp
}

// extractInterfaceMethods returns the list of methods declared
// directly on an interface type.
func extractInterfaceMethods(
	fset *token.FileSet, typ *doc.Type,
) []FunctionInfo {
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
			var tp string
			if field.Doc != nil {
				// field.Doc comes from the interface method declaration.
			}
			// Generate a signature and extract type parameters.
			sig, tp := interfaceMethodSignature(fset, field, name.Name)
			methodDoc := ""
			if field.Doc != nil {
				methodDoc = field.Doc.Text()
			} else if field.Comment != nil {
				methodDoc = field.Comment.Text()
			}
			methods = append(methods, FunctionInfo{
				Name:       name.Name,
				Doc:        methodDoc,
				Signature:  sig,
				TypeParams: tp,
			})
		}
	}
	return methods
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
		// Sort functions alphabetically.
		sort.Slice(apkg.Funcs, func(i, j int) bool {
			return apkg.Funcs[i].Name < apkg.Funcs[j].Name
		})
		for _, fn := range apkg.Funcs {
			fmt.Printf("### %s\n\n", fn.Name)
			// If function is generic, print type parameters.
			tpl := nodeToString(fset, fn.Decl.Type.TypeParams)
			if len(tpl) != 0 {
				fmt.Printf("**Type Parameters:** %s\n\n", tpl)
			}
			if fn.Doc != "" {
				fmt.Printf("%s\n\n", fn.Doc)
			}
			signature := nodeToString(fset, fn.Decl)
			fmt.Printf("```go\n%s\n```\n\n", signature)
		}
	}

	// Document only exported types.
	if len(apkg.Types) > 0 {
		fmt.Print("## Types\n\n")
		var types []*doc.Type
		// Filter out private types.
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
			// Obtain the TypeSpec from the declaration.
			var ts *ast.TypeSpec
			if typ.Decl != nil && len(typ.Decl.Specs) > 0 {
				ts, _ = typ.Decl.Specs[0].(*ast.TypeSpec)
			}
			// If the type is generic, display its type parameters.
			if ts != nil && ts.TypeParams != nil {
				tp := nodeToString(fset, ts.TypeParams)
				if len(tp) != 0 {
					fmt.Printf("**Type Parameters:** %s\n\n", tp)
				}
			}
			if typ.Doc != "" {
				fmt.Printf("%s\n\n", typ.Doc)
			}

			signature := nodeToString(fset, typ.Decl)
			if len(signature) < 80 {
				fmt.Printf("```go\n%s\n```\n\n", signature)
			}

			// Gather methods.
			methodMap := make(map[string]FunctionInfo)
			// For interface types, extract methods declared in the type literal.
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
				// Only include exported methods.
				if !ast.IsExported(m.Name) {
					continue
				}
				mSig := nodeToString(fset, m.Decl)
				tp := ""
				tp = nodeToString(fset, m.Decl.Type.TypeParams)
				methodMap[m.Name] = FunctionInfo{
					Name:       m.Name,
					Doc:        m.Doc,
					Signature:  mSig,
					TypeParams: tp,
				}
			}

			// If there are methods, document them.
			if len(methodMap) > 0 {
				fmt.Print("#### Methods\n\n")
				var mNames []string
				for name := range methodMap {
					mNames = append(mNames, name)
				}
				sort.Strings(mNames)
				for _, name := range mNames {
					m := methodMap[name]
					fmt.Printf("##### %s\n\n", m.Name)
					if m.TypeParams != "" {
						fmt.Printf("**Type Parameters:** %s\n\n", m.TypeParams)
					}
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
