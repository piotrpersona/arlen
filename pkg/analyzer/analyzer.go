package analyzer

import (
	"errors"
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const (
	lenFunctionName = "len"
	arlenCmd        = "arlen"
)

var Analyzer = &analysis.Analyzer{
	Name: arlenCmd,
	Doc:  "verifies if array len was checked before accessing the array",
	Run:  run,
}

func getTypesInfo(pass *analysis.Pass) (info *types.Info, err error) {
	conf := types.Config{Importer: importer.Default()}

	if pass.Pkg == nil {
		return nil, errors.New("package is nil")
	}

	info = &types.Info{
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	pkgName := pass.Pkg.Name()
	_, err = conf.Check(pkgName, pass.Fset, pass.Files, info)
	if err != nil {
		err = fmt.Errorf("cannot check package '%s, err: %w", pkgName, err)
	}
	return
}

func run(pass *analysis.Pass) (i interface{}, err error) {
	info, err := getTypesInfo(pass)
	if err != nil {
		return
	}

	arlenAnalyzer := NewArlenAnalyzer(pass, info)
	for _, f := range pass.Files {
		ast.Inspect(f, arlenAnalyzer.Inspect)
	}
	return
}

type ArlenAnalyzer struct {
	pass          *analysis.Pass
	objects       map[string]types.Object
	varLenChecked map[string]struct{}
}

func NewArlenAnalyzer(pass *analysis.Pass, info *types.Info) *ArlenAnalyzer {
	objects := make(map[string]types.Object, len(info.Defs))
	for ident, object := range info.Defs {
		objects[ident.Name] = object
	}
	return &ArlenAnalyzer{
		varLenChecked: make(map[string]struct{}, 0),
		pass:          pass,
		objects:       objects,
	}
}

func (a *ArlenAnalyzer) Inspect(node ast.Node) (procceed bool) {
	procceed = true

	switch stmt := node.(type) {
	case *ast.IndexExpr:
		a.verifyArrayCheck(stmt)
	case *ast.IfStmt:
		a.registerArrayCheck(stmt)
	}
	return
}

func (a *ArlenAnalyzer) verifyArrayCheck(expr *ast.IndexExpr) {
	ident, ok := expr.X.(*ast.Ident)
	if !ok {
		return
	}
	name := ident.Name
	object, ok := a.objects[name]
	if !ok {
		return
	}

	_, isSlice := object.Type().(*types.Slice)
	_, isArray := object.Type().(*types.Array)

	if !isSlice && !isArray {
		return
	}

	if _, ok := a.varLenChecked[identID(ident)]; ok {
		return
	}
	artype := "slice"
	if isArray {
		artype = "array"
	}
	a.report(ident.Pos(), "check %s %s length before accessing", artype, name)
}

func (a *ArlenAnalyzer) report(pos token.Pos, format string, args ...any) {
	a.pass.Reportf(pos, "%s: %s", arlenCmd, fmt.Sprintf(format, args...))
}

func (a *ArlenAnalyzer) registerArrayCheck(stmt *ast.IfStmt) {
	binaryExpr, ok := stmt.Cond.(*ast.BinaryExpr)
	if !ok {
		return
	}
	lenExpr := a.getLenCallExpr(binaryExpr)
	if lenExpr == nil {
		return
	}
	for _, arg := range lenExpr.Args {
		if ident, ok := arg.(*ast.Ident); ok {
			a.varLenChecked[identID(ident)] = struct{}{}
		}
	}
}

func (a *ArlenAnalyzer) getLenCallExpr(binaryExpr *ast.BinaryExpr) (expr *ast.CallExpr) {
	if expr := a.getLenExpr(binaryExpr.X); expr != nil {
		return expr
	}
	return a.getLenExpr(binaryExpr.Y)
}

func (a *ArlenAnalyzer) getLenExpr(expr ast.Expr) *ast.CallExpr {
	callExpr, ok := expr.(*ast.CallExpr)
	if !ok {
		return nil
	}
	if fnIdent, ok := callExpr.Fun.(*ast.Ident); ok && fnIdent.Name != lenFunctionName {
		return nil
	}
	return callExpr
}

func identID(ident *ast.Ident) string {
	return fmt.Sprintf("%s_%d", ident.Name, ident.Obj.Pos())
}

func exprID(expr ast.Expr) string {
	return fmt.Sprintf("%d_%d", expr.Pos(), expr.End())
}
