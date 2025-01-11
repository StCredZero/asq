package metaq

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestArrayType(t *testing.T) {
	src := `package test
func main() {
	//asq_start
	var x [5]int
	//asq_end
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	var arrayType *ast.ArrayType
	ast.Inspect(file, func(n ast.Node) bool {
		if at, ok := n.(*ast.ArrayType); ok {
			arrayType = at
			return false
		}
		return true
	})

	if arrayType == nil {
		t.Fatal("Failed to find ArrayType node")
	}

	var buf bytes.Buffer
	node := BuildAsqNode(arrayType, make(map[*ast.Ident]bool))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(array_type length: (literal) @value (#eq? @value "5") element: (identifier) @name (#eq? @name "int"))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestBasicLit(t *testing.T) {
	src := `package test
func main() {
	//asq_start
	x := 42
	//asq_end
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	var basicLit *ast.BasicLit
	ast.Inspect(file, func(n ast.Node) bool {
		if bl, ok := n.(*ast.BasicLit); ok {
			basicLit = bl
			return false
		}
		return true
	})

	if basicLit == nil {
		t.Fatal("Failed to find BasicLit node")
	}

	var buf bytes.Buffer
	node := BuildAsqNode(basicLit, make(map[*ast.Ident]bool))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(literal) @value (#eq? @value "42")`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestStructType(t *testing.T) {
	src := `package test
func main() {
	//asq_start
	type Person struct {
		Name string
		Age  int
	}
	//asq_end
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	var structType *ast.StructType
	ast.Inspect(file, func(n ast.Node) bool {
		if st, ok := n.(*ast.StructType); ok {
			structType = st
			return false
		}
		return true
	})

	if structType == nil {
		t.Fatal("Failed to find StructType node")
	}

	var buf bytes.Buffer
	node := BuildAsqNode(structType, make(map[*ast.Ident]bool))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(struct_type fields: (field_list (field_declaration names: ((identifier) @name (#eq? @name "Name")) type: (identifier) @name (#eq? @name "string")) (field_declaration names: ((identifier) @name (#eq? @name "Age")) type: (identifier) @name (#eq? @name "int"))))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestMapType(t *testing.T) {
	src := `package test
func main() {
	//asq_start
	var m map[string]int
	//asq_end
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	var mapType *ast.MapType
	ast.Inspect(file, func(n ast.Node) bool {
		if mt, ok := n.(*ast.MapType); ok {
			mapType = mt
			return false
		}
		return true
	})

	if mapType == nil {
		t.Fatal("Failed to find MapType node")
	}

	var buf bytes.Buffer
	node := BuildAsqNode(mapType, make(map[*ast.Ident]bool))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(map_type key: (identifier) @name (#eq? @name "string") value: (identifier) @name (#eq? @name "int"))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestFuncType(t *testing.T) {
	src := `package test
func main() {
	//asq_start
	type Handler func(string, int) bool
	//asq_end
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	var funcType *ast.FuncType
	ast.Inspect(file, func(n ast.Node) bool {
		if ft, ok := n.(*ast.FuncType); ok {
			funcType = ft
			return false
		}
		return true
	})

	if funcType == nil {
		t.Fatal("Failed to find FuncType node")
	}

	var buf bytes.Buffer
	node := BuildAsqNode(funcType, make(map[*ast.Ident]bool))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(function_type parameters: (field_list (field_declaration type: (identifier) @name (#eq? @name "string")) (field_declaration type: (identifier) @name (#eq? @name "int"))) results: (field_list (field_declaration type: (identifier) @name (#eq? @name "bool"))))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestGenDecl(t *testing.T) {
	src := `package test
//asq_start
const (
	A = 1
	B = 2
)
//asq_end`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	var genDecl *ast.GenDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if gd, ok := n.(*ast.GenDecl); ok {
			genDecl = gd
			return false
		}
		return true
	})

	if genDecl == nil {
		t.Fatal("Failed to find GenDecl node")
	}

	var buf bytes.Buffer
	node := BuildAsqNode(genDecl, make(map[*ast.Ident]bool))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(generic_declaration (value_spec names: ((identifier) @name (#eq? @name "A")) values: ((literal) @value (#eq? @value "1"))) (value_spec names: ((identifier) @name (#eq? @name "B")) values: ((literal) @value (#eq? @value "2"))))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestValueSpec(t *testing.T) {
	src := `package test
func main() {
	//asq_start
	var x, y int = 1, 2
	//asq_end
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	var valueSpec *ast.ValueSpec
	ast.Inspect(file, func(n ast.Node) bool {
		if vs, ok := n.(*ast.ValueSpec); ok {
			valueSpec = vs
			return false
		}
		return true
	})

	if valueSpec == nil {
		t.Fatal("Failed to find ValueSpec node")
	}

	var buf bytes.Buffer
	node := BuildAsqNode(valueSpec, make(map[*ast.Ident]bool))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(value_spec names: ((identifier) @name (#eq? @name "x") (identifier) @name (#eq? @name "y")) type: (identifier) @name (#eq? @name "int") values: ((literal) @value (#eq? @value "1") (literal) @value (#eq? @value "2")))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestBadDecl(t *testing.T) {
	badDecl := &ast.BadDecl{}
	var buf bytes.Buffer
	node := BuildAsqNode(badDecl, make(map[*ast.Ident]bool))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(bad_declaration)`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestChanType(t *testing.T) {
	src := `package test
func main() {
	//asq_start
	var ch chan int
	//asq_end
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	var chanType *ast.ChanType
	ast.Inspect(file, func(n ast.Node) bool {
		if ct, ok := n.(*ast.ChanType); ok {
			chanType = ct
			return false
		}
		return true
	})

	if chanType == nil {
		t.Fatal("Failed to find ChanType node")
	}

	var buf bytes.Buffer
	node := BuildAsqNode(chanType, make(map[*ast.Ident]bool))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(channel_type value: (identifier) @name (#eq? @name "int"))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestCompositeLit(t *testing.T) {
	src := `package test
func main() {
	//asq_start
	x := []int{1, 2}
	//asq_end
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	var compositeLit *ast.CompositeLit
	ast.Inspect(file, func(n ast.Node) bool {
		if cl, ok := n.(*ast.CompositeLit); ok {
			compositeLit = cl
			return false
		}
		return true
	})

	if compositeLit == nil {
		t.Fatal("Failed to find CompositeLit node")
	}

	var buf bytes.Buffer
	node := BuildAsqNode(compositeLit, make(map[*ast.Ident]bool))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(composite_literal type: (array_type element: (identifier) @name (#eq? @name "int")) elements: ((literal) @value (#eq? @value "1") (literal) @value (#eq? @value "2")))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestFuncDecl(t *testing.T) {
	src := `package test
//asq_start
func Add(x, y int) int {
	return x + y
}
//asq_end`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	var funcDecl *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok {
			funcDecl = fd
			return false
		}
		return true
	})

	if funcDecl == nil {
		t.Fatal("Failed to find FuncDecl node")
	}

	var buf bytes.Buffer
	node := BuildAsqNode(funcDecl, make(map[*ast.Ident]bool))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(function_declaration name: (identifier) @name (#eq? @name "Add") parameters: (field_list (field_declaration names: ((identifier) @name (#eq? @name "x") (identifier) @name (#eq? @name "y")) type: (identifier) @name (#eq? @name "int"))) results: (field_list (field_declaration type: (identifier) @name (#eq? @name "int"))) body: (block (return_statement values: (expression_list (binary_expression left: (identifier) @name (#eq? @name "x") operator: "+" right: (identifier) @name (#eq? @name "y"))))))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestFuncLit(t *testing.T) {
	src := `package test
func main() {
	//asq_start
	f := func(x int) int { return x * 2 }
	//asq_end
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	var funcLit *ast.FuncLit
	ast.Inspect(file, func(n ast.Node) bool {
		if fl, ok := n.(*ast.FuncLit); ok {
			funcLit = fl
			return false
		}
		return true
	})

	if funcLit == nil {
		t.Fatal("Failed to find FuncLit node")
	}

	var buf bytes.Buffer
	node := BuildAsqNode(funcLit, make(map[*ast.Ident]bool))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(function_literal type: (function_type parameters: (field_list (field_declaration names: ((identifier) @name (#eq? @name "x")) type: (identifier) @name (#eq? @name "int"))) results: (field_list (field_declaration type: (identifier) @name (#eq? @name "int")))))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestPackage(t *testing.T) {
	src := `//asq_start
package test
//asq_end`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	var buf bytes.Buffer
	node := BuildAsqNode(file, make(map[*ast.Ident]bool))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(source_file package_name: (identifier) @name (#eq? @name "test"))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestTypeSpec(t *testing.T) {
	src := `package test
//asq_start
type Point struct {
	X, Y int
}
//asq_end`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	var typeSpec *ast.TypeSpec
	ast.Inspect(file, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok {
			typeSpec = ts
			return false
		}
		return true
	})

	if typeSpec == nil {
		t.Fatal("Failed to find TypeSpec node")
	}

	var buf bytes.Buffer
	node := BuildAsqNode(typeSpec, make(map[*ast.Ident]bool))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(type_spec name: (identifier) @name (#eq? @name "Point") type: (struct_type fields: (field_list (field_declaration names: ((identifier) @name (#eq? @name "X") (identifier) @name (#eq? @name "Y")) type: (identifier) @name (#eq? @name "int")))))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}
