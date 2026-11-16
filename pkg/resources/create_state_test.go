package resources_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// creationCall matches the calls that make the remote object exist: SDK
// operations (Createprovider, Provisionaddonbyorgaid), the helpers wrapping
// them (CreateAddon, application.Create) and the S3 bucket creation.
var creationCall = regexp.MustCompile(`^(Create|Provision|MakeBucket)`)

// Once the API accepted a creation, the object exists and is billed: Create has
// to save it in state before anything else can fail, otherwise Terraform loses
// track of it, the next apply creates a second one and destroy leaves it behind.
//
// This reads the Create method of every resource and refuses a `return` placed
// between the creation call (and the check of its own error) and the first
// State.Set.
func TestCreateSavesStateRightAfterCreation(t *testing.T) {
	fset := token.NewFileSet()
	found := 0

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !isResourceCreate(fn) {
				continue
			}
			found++

			for _, problem := range checkCreate(fn) {
				t.Errorf("%s: %s", fset.Position(problem.pos), problem.msg)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// drains share one Create, java, scala, static, play2... have their own
	if found < 30 {
		t.Fatalf("found %d Create methods only, the walk is broken", found)
	}
}

// isResourceCreate reports whether fn is the Create method of a resource.Resource
func isResourceCreate(fn *ast.FuncDecl) bool {
	if fn.Name.Name != "Create" || fn.Recv == nil || fn.Body == nil {
		return false
	}
	for _, param := range fn.Type.Params.List {
		if sel, ok := param.Type.(*ast.SelectorExpr); ok && sel.Sel.Name == "CreateRequest" {
			return true
		}
	}
	return false
}

type problem struct {
	pos token.Pos
	msg string
}

func checkCreate(fn *ast.FuncDecl) []problem {
	// the top-level statement holding the creation call
	creationIdx := -1
	var creationEnd token.Pos
	for i, stmt := range fn.Body.List {
		if end := firstCall(stmt, isCreationCall); end.IsValid() {
			creationIdx, creationEnd = i, end
			break
		}
	}
	if creationIdx == -1 {
		return []problem{{fn.Pos(), "no creation call recognised in Create: name it in creationCall so that the state save can be checked"}}
	}

	firstSave := firstCall(fn.Body, func(call *ast.CallExpr) bool {
		return call.Pos() > creationEnd && isStateSet(call)
	})
	if !firstSave.IsValid() {
		return []problem{{fn.Pos(), "Create never saves the state after the creation call"}}
	}

	// the statement right after the creation is the check of its own error:
	// nothing was created when it returns
	allowedEnd := fn.Body.List[creationIdx].End()
	if creationIdx+1 < len(fn.Body.List) {
		allowedEnd = fn.Body.List[creationIdx+1].End()
	}

	var problems []problem
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.ReturnStmt:
			if node.Pos() > allowedEnd && node.Pos() < firstSave {
				problems = append(problems, problem{node.Pos(), "return after the creation call and before the first State.Set: the created object would be lost, save the state first"})
			}
		}
		return true
	})
	return problems
}

// firstCall returns the end of the first call matching, in source order
func firstCall(root ast.Node, match func(*ast.CallExpr) bool) token.Pos {
	found := token.NoPos
	ast.Inspect(root, func(n ast.Node) bool {
		if found.IsValid() {
			return false
		}
		if call, ok := n.(*ast.CallExpr); ok && match(call) {
			found = call.End()
			return false
		}
		return true
	})
	return found
}

func isCreationCall(call *ast.CallExpr) bool {
	fun := call.Fun
	// generic instantiation: application.Create[T](...)
	if idx, ok := fun.(*ast.IndexExpr); ok {
		fun = idx.X
	}
	switch f := fun.(type) {
	case *ast.SelectorExpr:
		return creationCall.MatchString(f.Sel.Name)
	case *ast.Ident:
		return creationCall.MatchString(f.Name)
	}
	return false
}

// isStateSet matches <response>.State.Set(...)
func isStateSet(call *ast.CallExpr) bool {
	set, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || set.Sel.Name != "Set" {
		return false
	}
	state, ok := set.X.(*ast.SelectorExpr)
	return ok && state.Sel.Name == "State"
}
