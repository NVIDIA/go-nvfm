// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES.  All rights reserved.
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

type generatableInterface struct {
	Type                      string
	Interface                 string
	Exclude                   []string
	PackageMethodsAliasedFrom string
}

var generatableInterfaces = []generatableInterface{
	{
		Type:                      "library",
		Interface:                 "Interface",
		Exclude:                   []string{"LookupSymbol"},
		PackageMethodsAliasedFrom: "libnvfm",
	},
	{
		Type:      "fabricManager",
		Interface: "Handle",
	},
}

func main() {
	sourceDir := flag.String("sourceDir", "", "Path to the source directory for Go files")
	output := flag.String("output", "", "Path to the output file (default: stdout)")
	flag.Parse()

	if *sourceDir == "" {
		flag.Usage()
		os.Exit(2)
	}

	writer, closer, err := getWriter(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening output: %v\n", err)
		os.Exit(1)
	}
	defer closer()

	var buf bytes.Buffer
	buf.WriteString(header())

	for i, iface := range generatableInterfaces {
		methods, err := extractMethods(*sourceDir, iface)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error extracting methods: %v\n", err)
			os.Exit(1)
		}

		if iface.PackageMethodsAliasedFrom != "" {
			buf.WriteString(packageMethods(iface, methods))
			buf.WriteString("\n")
		}

		buf.WriteString(interfaceDefinition(iface, methods))
		if i < len(generatableInterfaces)-1 {
			buf.WriteString("\n")
		}
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error formatting output: %v\n%s", err, buf.String())
		os.Exit(1)
	}
	if _, err := writer.Write(formatted); err != nil {
		fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
		os.Exit(1)
	}
}

func getWriter(outputFile string) (io.Writer, func() error, error) {
	if outputFile == "" {
		return os.Stdout, func() error { return nil }, nil
	}
	file, err := os.Create(outputFile)
	if err != nil {
		return nil, nil, err
	}
	return file, file.Close, nil
}

func header() string {
	return `// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES.  All rights reserved.
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Generated Code; DO NOT EDIT.

package nvfm

`
}

func packageMethods(iface generatableInterface, methods []*ast.FuncDecl) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "// The variables below represent package-level methods from the %s type.\n", iface.Type)
	buf.WriteString("var (\n")
	for _, method := range methods {
		fmt.Fprintf(&buf, "\t%s = %s.%s\n", method.Name.Name, iface.PackageMethodsAliasedFrom, method.Name.Name)
	}
	buf.WriteString(")\n")
	return buf.String()
}

func interfaceDefinition(iface generatableInterface, methods []*ast.FuncDecl) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "// %s represents the interface for the %s type.\n", iface.Interface, iface.Type)
	fmt.Fprintf(&buf, "type %s interface {\n", iface.Interface)
	for _, method := range methods {
		fmt.Fprintf(&buf, "\t%s%s%s\n", method.Name.Name, fieldList(method.Type.Params), results(method.Type.Results))
	}
	buf.WriteString("}\n")
	return buf.String()
}

func extractMethods(sourceDir string, iface generatableInterface) ([]*ast.FuncDecl, error) {
	fset := token.NewFileSet()
	var methods []*ast.FuncDecl

	err := filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		base := filepath.Base(path)
		if base == "zz_generated.api.go" || base == "types.go" || strings.HasSuffix(base, "_test.go") || strings.HasSuffix(base, "_mock.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}
			if receiverType(fn.Recv.List[0].Type) != iface.Type {
				continue
			}
			if !fn.Name.IsExported() {
				continue
			}
			if slices.Contains(iface.Exclude, fn.Name.Name) {
				continue
			}
			methods = append(methods, fn)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(methods, func(i, j int) bool {
		return methods[i].Name.Name < methods[j].Name.Name
	})
	return methods, nil
}

func receiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return receiverType(t.X)
	default:
		return ""
	}
}

func fieldList(fields *ast.FieldList) string {
	if fields == nil {
		return "()"
	}
	parts := make([]string, 0, len(fields.List))
	for _, field := range fields.List {
		typ := formatNode(field.Type)
		if len(field.Names) == 0 {
			parts = append(parts, typ)
			continue
		}
		names := make([]string, 0, len(field.Names))
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
		parts = append(parts, strings.Join(names, ", ")+" "+typ)
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func results(fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}
	if len(fields.List) == 1 && len(fields.List[0].Names) == 0 {
		return " " + formatNode(fields.List[0].Type)
	}
	return " " + fieldList(fields)
}

func formatNode(node any) string {
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, token.NewFileSet(), node)
	return buf.String()
}
