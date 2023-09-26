package enums

import (
	"flag"
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"reflect"
	"strconv"

	"golang.org/x/tools/go/packages"
)

func Get(ty reflect.Type) ([]any, error) {
	var enum []any
	config := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		// Only look in test files if a test is in progress.
		Tests: flag.Lookup("test.v") != nil,
	}
	config.Fset = token.NewFileSet()
	pkgs, err := packages.Load(config, ty.PkgPath())
	if err != nil {
		return nil, fmt.Errorf("could not load package %q", ty.PkgPath())
	}
	for _, p := range pkgs {
		for _, syn := range p.Syntax {
			for _, d := range syn.Decls {
				if _, ok := d.(*ast.GenDecl); !ok {
					continue
				}
				for _, sp := range d.(*ast.GenDecl).Specs {
					v, ok := sp.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, name := range v.Names {
						v, err := getConstantValue(ty, name, p)
						if err != nil {
							return nil, err
						}
						if v != nil {
							enum = append(enum, v)
						}
					}
				}
			}
		}
	}
	return enum, nil
}

func getConstantValue(ty reflect.Type, name *ast.Ident, pkg *packages.Package) (any, error) {
	c, ok := pkg.TypesInfo.ObjectOf(name).(*types.Const)
	if !ok {
		return nil, nil
	}
	if c.Type().String() == ty.PkgPath()+"."+ty.Name() {
		if c.Val().Kind() == constant.String {
			return constant.StringVal(c.Val()), nil
		}
		if c.Val().Kind() == constant.Int {
			n, err := strconv.Atoi(c.Val().ExactString())
			if err != nil {
				return nil, fmt.Errorf("could not parse enum %s value: %q", ty.Name(), c.Val().ExactString())
			}
			return n, nil
		}
		return c.Val().ExactString(), nil
	}
	return nil, nil
}
