package main

import (
	"bytes"
	"go/ast"
	"go/doc"
	"go/printer"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

var echoFiles = make(map[string]*os.File)

func echoFile(path string) *os.File {
	if f, ok := echoFiles[path]; ok {
		return f
	}
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil && !os.IsExist(err) {
		log.Fatalf("Failed to create directory: %v", err)
	}

	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Failed to create echo file: %v", err)
	}
	echoFiles[path] = f
	return f
}

func echo(text string, destination ...string) {
	f := echoFile(filepath.Join(destination...))

	_, err := f.WriteString(text)
	if err != nil {
		log.Fatalf("Failed to write to echo file: %v", err)
	}
}

func echoClose() {
	for _, f := range echoFiles {
		_ = f.Close()
	}
}

func md(doc []byte) string {
	pattern := `\t.+\n(?m:\t.+\n|\n\t.+\n)+`
	re := regexp.MustCompile(pattern)

	return re.ReplaceAllStringFunc(string(doc), func(block string) string {
		// Remove the four-space prefix from each line.
		lines := strings.Split(block, "\n")
		for i, line := range lines {
			line = strings.TrimPrefix(line, "\t")
			lines[i] = line
		}
		// Wrap the result in a fenced code block with "go"
		return "```go\n" + strings.Join(lines, "\n") + "```"
	})
}

func clearDir(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, e := range entries {
		if e.Name() == ".git" {
			continue
		}
		_ = os.RemoveAll(filepath.Join(dir, e.Name()))
	}
}

func writeFooter(destination ...string) {
	f := echoFile(filepath.Join(destination...))

	if licenseF, err := os.Open("LICENSE"); err == nil {
		_, _ = io.Copy(f, licenseF)
		_ = licenseF.Close()
	} else {
		_, _ = f.WriteString("All rights reserved")
	}
}
