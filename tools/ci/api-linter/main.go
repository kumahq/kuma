package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"os"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
	"golang.org/x/tools/go/packages"

	"github.com/kumahq/kuma/tools/ci/api-linter/linter"
)

func main() {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Dir:  ".",
		Fset: token.NewFileSet(),
	}

	entries, _ := os.ReadDir("api/common/v1alpha1")
	commonPackages := []string{"github.com/kumahq/kuma/api/common/v1alpha1"}
	for _, entry := range entries {
		if entry.IsDir() {
			commonPackages = append(commonPackages, "github.com/kumahq/kuma/api/common/v1alpha1/"+entry.Name())
		}
	}

	pkgs, err := packages.Load(cfg, commonPackages...)
	if err != nil {
		log.Fatalf("Failed to load packages: %v", err)
	}

	for _, pkg := range pkgs {
		pass := &analysis.Pass{
			Analyzer:  linter.Analyzer,
			Fset:      pkg.Fset,
			Pkg:       pkg.Types,
			Files:     pkg.Syntax,
			TypesInfo: pkg.TypesInfo,
			ResultOf:  map[*analysis.Analyzer]interface{}{},
			Report:    func(diag analysis.Diagnostic) { fmt.Println(diag) },
		}

		for _, file := range pass.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				typeSpec, ok := n.(*ast.TypeSpec)
				if !ok {
					return true
				}
				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					linter.CommonTypes[pkg.PkgPath+"."+typeSpec.Name.Name] = structType
				}
				return true
			})
		}
	}

	singlechecker.Main(linter.Analyzer)
}
