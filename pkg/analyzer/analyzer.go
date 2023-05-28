package analyzer

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const lenFunctionName = "len"

var Analyzer = &analysis.Analyzer{
	Name: "arlen",
	Doc:  "verifies if array len was checked before accessing the array",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	conf := types.Config{Importer: importer.Default()}

	info := &types.Info{
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	_, err := conf.Check(pass.Pkg.Name(), pass.Fset, pass.Files, info)
	if err != nil {
		return nil, err
	}
	objects := make(map[string]types.Object, len(info.Defs))
	for ident, object := range info.Defs {
		objects[ident.Name] = object
	}

	arlenAnalyzer := NewArlenAnalyzer(pass, objects)
	for _, f := range pass.Files {
		ast.Inspect(f, arlenAnalyzer.Inspect)
	}
	return nil, nil
}

type ArlenAnalyzer struct {
	pass          *analysis.Pass
	objects       map[string]types.Object
	varLenChecked map[string]struct{}
}

func NewArlenAnalyzer(pass *analysis.Pass, objects map[string]types.Object) *ArlenAnalyzer {
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
		a.verifyIfVariableWasChecked(stmt)
	case *ast.IfStmt:
		a.registerCheckedVariables(stmt)
	}
	return
}

func (a *ArlenAnalyzer) verifyIfVariableWasChecked(expr *ast.IndexExpr) {
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

	if _, ok := a.varLenChecked[a.identKey(ident)]; ok {
		return
	}
	a.pass.Reportf(ident.Pos(), "arlen: check variable %s before accessing", name)
}

func (a *ArlenAnalyzer) registerCheckedVariables(stmt *ast.IfStmt) {
	binaryExpr, ok := stmt.Cond.(*ast.BinaryExpr)
	if !ok {
		return
	}
	lenExpr := a.getLenExpression(binaryExpr)
	if lenExpr == nil {
		return
	}
	for _, arg := range lenExpr.Args {
		if ident, ok := arg.(*ast.Ident); ok {
			a.varLenChecked[a.identKey(ident)] = struct{}{}
		}
	}
}

func (a *ArlenAnalyzer) getLenExpression(binaryExpr *ast.BinaryExpr) (expr *ast.CallExpr) {
	getCallExpr := func(expr ast.Expr) *ast.CallExpr {
		callExpr, ok := expr.(*ast.CallExpr)
		if !ok {
			return nil
		}
		if fnIdent, ok := callExpr.Fun.(*ast.Ident); ok && fnIdent.Name != lenFunctionName {
			return nil
		}
		return callExpr
	}

	if expr := getCallExpr(binaryExpr.X); expr != nil {
		return expr
	}
	return getCallExpr(binaryExpr.Y)
}

func (a *ArlenAnalyzer) identKey(ident *ast.Ident) string {
	return fmt.Sprintf("%s_%d", ident.Name, ident.Obj.Pos())
}
