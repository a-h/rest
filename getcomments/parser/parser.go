package parser

import (
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

var ErrGoCommandRequired = errors.New("go command required, not found")

func Get(packageName string) (m map[string]string, err error) {
	config := &packages.Config{
		Mode:  packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		Tests: true,
	}
	pkgs, err := packages.Load(config, packageName)
	if err != nil {
		if strings.HasPrefix(err.Error(), "err: go command required, not found: ") {
			err = ErrGoCommandRequired
		}
		err = fmt.Errorf("error loading package %s: %w", packageName, err)
		return
	}

	// Add the comments to the definitions.
	m = make(map[string]string)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			processFile(packageName, pkg, file, m)
		}
	}
	return
}

func processFile(packageName string, pkg *packages.Package, file *ast.File, m map[string]string) {
	var lastComment string
	var typ string
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			typ = x.Name.String()
			if !ast.IsExported(typ) {
				break
			}
			typeID := fmt.Sprintf("%s.%s", packageName, typ)
			if lastComment != "" {
				m[typeID] = lastComment
			}
		case *ast.GenDecl:
			lastComment = strings.TrimSpace(x.Doc.Text())
		case *ast.ValueSpec:
			// Get comments on constants, since they may appear in string and integer enums.
			for _, name := range x.Names {
				c, isConstant := pkg.TypesInfo.ObjectOf(name).(*types.Const)
				if !isConstant {
					continue
				}
				typeID := fmt.Sprintf("%s.%s", packageName, c.Name())
				comments := lastComment
				if strings.TrimSpace(x.Doc.Text()) != "" {
					comments = strings.TrimSpace(x.Doc.Text())
				}
				if comments != "" {
					m[typeID] = comments
				}
			}
		case *ast.FuncDecl:
			// Skip functions, since they can't appear in schema.
			return false
		case *ast.Field:
			if typ == "" {
				break
			}
			if !ast.IsExported(typ) {
				break
			}
			fieldName := getFieldName(x)
			if !ast.IsExported(fieldName) {
				break
			}
			typeID := fmt.Sprintf("%s.%s.%s", packageName, typ, fieldName)
			comments := strings.TrimSpace(x.Doc.Text())
			if comments != "" {
				m[typeID] = comments
			}
		}
		return true
	})
}

func getFieldName(field *ast.Field) string {
	var names []string
	for _, name := range field.Names {
		names = append(names, name.Name)
	}
	return strings.Join(names, ".")
}
