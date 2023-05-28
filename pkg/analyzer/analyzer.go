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

	SlenCmd         = "slen"
	SlenDescription = "verifies if slice len was checked before accessing the array"
)

var Analyzer = &analysis.Analyzer{
	Name: SlenCmd,
	Doc:  SlenDescription,
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

	slenAnalyzer := NewSlenAnalyzer(pass, info)
	for _, f := range pass.Files {
		ast.Inspect(f, slenAnalyzer.Inspect)
	}
	return
}

type SlenAnalyzer struct {
	pass          *analysis.Pass
	objects       map[string]types.Object
	varLenChecked map[string]struct{}
}

func NewSlenAnalyzer(pass *analysis.Pass, info *types.Info) *SlenAnalyzer {
	objects := make(map[string]types.Object, len(info.Defs))
	for ident, object := range info.Defs {
		objects[ident.Name] = object
	}
	return &SlenAnalyzer{
		varLenChecked: make(map[string]struct{}, 0),
		pass:          pass,
		objects:       objects,
	}
}

func (a *SlenAnalyzer) Inspect(node ast.Node) (procceed bool) {
	switch stmt := node.(type) {
	// verify
	case *ast.IndexExpr:
		a.verifyIndexExpr(stmt)
	case *ast.SliceExpr:
		a.verifySliceExpr(stmt)

	// register
	case *ast.IfStmt:
		a.registerIfStmt(stmt)
	case *ast.RangeStmt:
		a.registerRangeStmt(stmt)
	case *ast.ForStmt:
		a.registerForStmt(stmt)
	}
	return true
}

func (a *SlenAnalyzer) verifySliceExpr(stmt *ast.SliceExpr) {
	if ident, ok := stmt.X.(*ast.Ident); ok && !a.sliceLengthChecked(ident) {
		a.report(ident.Pos(), "check slice %s length before accessing", ident.Name)
	}
}

func (a *SlenAnalyzer) verifyIndexExpr(expr *ast.IndexExpr) {
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
	if !isSlice {
		return
	}

	if a.sliceLengthChecked(ident) {
		return
	}

	a.report(ident.Pos(), "check slice %s length before accessing", name)
}

func (a *SlenAnalyzer) registerCheck(ident *ast.Ident) {
	a.varLenChecked[identID(ident)] = struct{}{}
}

func (a *SlenAnalyzer) sliceLengthChecked(ident *ast.Ident) (checked bool) {
	_, checked = a.varLenChecked[identID(ident)]
	return
}

func (a *SlenAnalyzer) registerForStmt(stmt *ast.ForStmt) {
	if binaryExpr, ok := stmt.Cond.(*ast.BinaryExpr); ok {
		a.registerCondCheck(binaryExpr)
	}
}

func (a *SlenAnalyzer) registerRangeStmt(stmt *ast.RangeStmt) {
	if ident, ok := stmt.X.(*ast.Ident); ok {
		a.registerCheck(ident)
	}
}

func (a *SlenAnalyzer) registerIfStmt(stmt *ast.IfStmt) {
	if binaryExpr, ok := stmt.Cond.(*ast.BinaryExpr); ok {
		a.registerCondCheck(binaryExpr)
	}
}

func (a *SlenAnalyzer) registerCondCheck(binaryExpr *ast.BinaryExpr) {
	lenExpr := a.getLenCallExpr(binaryExpr)
	if lenExpr == nil {
		return
	}
	for _, arg := range lenExpr.Args {
		if ident, ok := arg.(*ast.Ident); ok {
			a.registerCheck(ident)
		}
	}
}

func (a *SlenAnalyzer) getLenCallExpr(binaryExpr *ast.BinaryExpr) (expr *ast.CallExpr) {
	if expr := a.getLenExpr(binaryExpr.X); expr != nil {
		return expr
	}
	return a.getLenExpr(binaryExpr.Y)
}

func (a *SlenAnalyzer) getLenExpr(expr ast.Expr) *ast.CallExpr {
	callExpr, ok := expr.(*ast.CallExpr)
	if !ok {
		return nil
	}
	if fnIdent, ok := callExpr.Fun.(*ast.Ident); ok && fnIdent.Name != lenFunctionName {
		return nil
	}
	return callExpr
}

func (a *SlenAnalyzer) report(pos token.Pos, format string, args ...any) {
	a.pass.Reportf(pos, "%s: %s", SlenCmd, fmt.Sprintf(format, args...))
}

func identID(ident *ast.Ident) string {
	return fmt.Sprintf("%s_%d", ident.Name, ident.Obj.Pos())
}
