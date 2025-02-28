package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"log"
	"os"
	"sort"
	"strings"
)

var (
	outDir = flag.String("o", "./docs", "output directly for generated documentation")
)

func main() {
	flag.Parse()

	defer echoClose()

	if err := os.MkdirAll(*outDir, 0777); err != nil && !os.IsExist(err) {
		log.Fatalf("Error creating output directly: %v", err)
	}

	clearDir(*outDir)

	fset := token.NewFileSet()
	// Parse the current directory with comments.
	pkgs, err := parser.ParseDir(fset, ".", nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Error parsing directory: %v", err)
	}

	// Use the first package found.
	var apkg *doc.Package
	for _, pkg := range pkgs {
		if strings.HasSuffix(pkg.Name, "_test") {
			continue
		}
		// The "./" import path is just a placeholder.
		apkg = doc.New(pkg, "./", doc.AllDecls)
		break
	}
	if apkg == nil {
		log.Fatal("No package found")
	}

	// Package header.
	echo(fmt.Sprintf("# Package %s\n\n", apkg.Name), *outDir, "Home.md")
	echo(fmt.Sprintf("- [%s](Home)\n", apkg.Name), *outDir, "_Sidebar.md")
	if apkg.Doc != "" {
		echo(fmt.Sprintf("%s\n\n", md(apkg.Markdown(apkg.Doc))), *outDir, "Home.md")
	}

	// Document package-level functions.
	if len(apkg.Funcs) > 0 {
		echo("- Functions\n", *outDir, "_Sidebar.md")
		// Sort functions alphabetically.
		sort.Slice(apkg.Funcs, func(i, j int) bool {
			return apkg.Funcs[i].Name < apkg.Funcs[j].Name
		})
		for _, fn := range apkg.Funcs {
			if !ast.IsExported(fn.Name) {
				continue
			}
			echo(fmt.Sprintf("  - [%s](%s)\n", fn.Name, fn.Name), *outDir, "_Sidebar.md")
			// If function is generic, print type parameters.
			tpl := nodeToString(fset, fn.Decl.Type.TypeParams)
			if len(tpl) != 0 {
				echo(fmt.Sprintf("**Type Parameters:** %s\n\n", tpl), *outDir, "func", fn.Name+".md")
			}
			if fn.Doc != "" {
				echo(fmt.Sprintf("%s\n\n", md(apkg.Markdown(fn.Doc))), *outDir, "func", fn.Name+".md")
			}
			signature := nodeToString(fset, fn.Decl)
			echo(fmt.Sprintf("```go\n%s\n```\n\n", signature), *outDir, "func", fn.Name+".md")
		}
	}

	// Document only exported types.
	if len(apkg.Types) > 0 {
		echo("- Types\n", *outDir, "_Sidebar.md")
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
			echo(fmt.Sprintf("  - [%s](%s)\n", typ.Name, typ.Name), *outDir, "_Sidebar.md")
			// Obtain the TypeSpec from the declaration.
			var ts *ast.TypeSpec
			if typ.Decl != nil && len(typ.Decl.Specs) > 0 {
				ts, _ = typ.Decl.Specs[0].(*ast.TypeSpec)
			}
			// If the type is generic, display its type parameters.
			if ts != nil && ts.TypeParams != nil {
				tp := nodeToString(fset, ts.TypeParams)
				if len(tp) != 0 {
					echo(fmt.Sprintf("**Type Parameters:** %s\n\n", tp), *outDir, "type", typ.Name+".md")
				}
			}
			if typ.Doc != "" {
				echo(fmt.Sprintf("%s\n\n", typ.Doc), *outDir, "type", typ.Name+".md")
			}

			signature := nodeToString(fset, typ.Decl)
			if len(signature) < 80 {
				echo(fmt.Sprintf("```go\n%s\n```\n\n", signature), *outDir, "type", typ.Name+".md")
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
				echo("## Methods\n\n", *outDir, "type", typ.Name+".md")
				var mNames []string
				for name := range methodMap {
					mNames = append(mNames, name)
				}
				sort.Strings(mNames)
				for _, name := range mNames {
					m := methodMap[name]
					echo(fmt.Sprintf("### %s\n\n", m.Name), *outDir, "type", typ.Name+".md")
					if m.TypeParams != "" {
						echo(fmt.Sprintf("**Type Parameters:** %s\n\n", m.TypeParams), *outDir, "type", typ.Name+".md")
					}
					if m.Doc != "" {
						echo(fmt.Sprintf("%s\n\n", m.Doc), *outDir, "type", typ.Name+".md")
					}
					if m.Signature != "" {
						echo(fmt.Sprintf("```go\n%s\n```\n\n", m.Signature), *outDir, "type", typ.Name+".md")
					}
				}
			}
		}
	}

	writeFooter(*outDir, "_Footer.md")
}
