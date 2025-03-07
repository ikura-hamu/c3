package c3

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/ssa"
)

const doc = "testing_cleanup_ctx is an analyzer to detect calling (*testing.common).Context inside (*testing.common).Cleanup"

// Analyzer is an analyzer to detect calling (*testing.common).Context inside (*testing.common).Cleanup
var Analyzer = &analysis.Analyzer{
	Name: "testing_cleanup_ctx",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
		buildssa.Analyzer,
	},
}

func run(pass *analysis.Pass) (any, error) {
	cleanupFuncPositions := make(map[token.Pos]bool)
	inspec := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Find all Cleanup calls that take a function literal as an argument
	inspec.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}
		sel, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Cleanup" {
			return
		}
		obj := pass.TypesInfo.ObjectOf(sel.Sel)
		if obj == nil {
			return
		}
		funcObj, ok := obj.(*types.Func)
		if !ok {
			return
		}
		sig, ok := funcObj.Type().(*types.Signature)
		if !ok || sig.Recv() == nil {
			return
		}
		recvType := sig.Recv().Type()
		ptr, ok := recvType.(*types.Pointer)
		if !ok {
			return
		}
		named, ok := ptr.Elem().(*types.Named)
		if !ok {
			return
		}
		if named.Obj().Name() != "common" || named.Obj().Pkg().Path() != "testing" {
			return
		}
		if len(callExpr.Args) == 0 {
			return
		}
		if funcLit, ok := callExpr.Args[0].(*ast.FuncLit); ok {
			cleanupFuncPositions[funcLit.Pos()] = true
		}
	})

	ssaResult := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)

	// Check if the Cleanup function calls (*testing.common).Context
	for _, fn := range ssaResult.SrcFuncs {
		if fn == nil || fn.Syntax() == nil {
			continue
		}
		if funcLit, ok := fn.Syntax().(*ast.FuncLit); ok {
			if !cleanupFuncPositions[funcLit.Pos()] {
				continue
			}
			visited := make(map[*ssa.Function]bool)
			if containsContextCall(fn, visited) {
				pass.ReportRangef(funcLit, "avoid calling (*testing.common).Context inside Cleanup")
			}
		}
	}

	return nil, nil
}

// isTestingCommonContext checks if the function is (*testing.common).Context
func isTestingCommonContext(fn *ssa.Function) bool {
	if fn.Name() != "Context" {
		return false
	}
	if fn.Signature.Recv() == nil {
		return false
	}
	recvType := fn.Signature.Recv().Type()
	ptr, ok := recvType.(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := ptr.Elem().(*types.Named)
	if !ok {
		return false
	}
	return named.Obj().Name() == "common" && named.Obj().Pkg().Path() == "testing"
}

// containsContextCall checks if the function contains a call to (*testing.common).Context
func containsContextCall(fn *ssa.Function, visited map[*ssa.Function]bool) bool {
	if visited[fn] {
		return false
	}
	visited[fn] = true
	for _, b := range fn.Blocks {
		for _, instr := range b.Instrs {
			callInstr, ok := instr.(ssa.CallInstruction)
			if !ok {
				continue
			}
			callee := callInstr.Common().StaticCallee()
			if callee == nil {
				continue
			}
			if isTestingCommonContext(callee) {
				return true
			}
			if containsContextCall(callee, visited) {
				return true
			}
		}
	}
	return false
}
