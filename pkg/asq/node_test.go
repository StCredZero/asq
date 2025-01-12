package asq

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestNonIdentWildcardTagging(t *testing.T) {
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

	p := newPassOne(file)
	var buf bytes.Buffer
	node := BuildAsqNode(basicLit, p)
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	// Verify that non-Ident nodes cannot be wildcarded
	expected := `(literal) (#eq? _ "42")`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestArrayType(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
func main() {
	//asq_start
	var x [5]int
	//asq_end
}`
	var err error
	var file *ast.File
	file, err = parser.ParseFile(fset, "", src, parser.ParseComments)
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
	var node Node
	node = BuildAsqNode(arrayType, newPassOne(file))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(array_type length: (literal) (#eq? _ "5") element: (identifier) (#eq? _ "int"))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestBasicLit(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
func main() {
	//asq_start
	x := 42
	//asq_end
}`
	var err error
	var file *ast.File
	file, err = parser.ParseFile(fset, "", src, parser.ParseComments)
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
	var node Node
	node = BuildAsqNode(basicLit, newPassOne(file))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(literal) (#eq? _ "42")`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestStructType(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
func main() {
	//asq_start
	type Person struct {
		Name string
		Age  int
	}
	//asq_end
}`
	var err error
	var file *ast.File
	file, err = parser.ParseFile(fset, "", src, parser.ParseComments)
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
	var node Node
	node = BuildAsqNode(structType, newPassOne(file))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(struct_type fields: (field_list (field_declaration names: ((identifier) (#eq? _ "Name")) type: (identifier) (#eq? _ "string")) (field_declaration names: ((identifier) (#eq? _ "Age")) type: (identifier) (#eq? _ "int"))))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestMapType(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
func main() {
	//asq_start
	var m map[string]int
	//asq_end
}`
	var err error
	var file *ast.File
	file, err = parser.ParseFile(fset, "", src, parser.ParseComments)
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
	var node Node
	node = BuildAsqNode(mapType, newPassOne(file))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(map_type key: (identifier) (#eq? _ "string") value: (identifier) (#eq? _ "int"))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestFuncType(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
func main() {
	//asq_start
	type Handler func(string, int) bool
	//asq_end
}`
	var err error
	var file *ast.File
	file, err = parser.ParseFile(fset, "", src, parser.ParseComments)
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
	var node Node
	node = BuildAsqNode(funcType, newPassOne(file))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(function_type parameters: (field_list (field_declaration type: (identifier) (#eq? _ "string")) (field_declaration type: (identifier) (#eq? _ "int"))) results: (field_list (field_declaration type: (identifier) (#eq? _ "bool"))))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestGenDecl(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
//asq_start
const (
	A = 1
	B = 2
)
//asq_end`
	var err error
	var file *ast.File
	file, err = parser.ParseFile(fset, "", src, parser.ParseComments)
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
	var node Node
	node = BuildAsqNode(genDecl, newPassOne(file))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(generic_declaration (value_spec names: ((identifier) (#eq? _ "A")) values: ((literal) (#eq? _ "1"))) (value_spec names: ((identifier) (#eq? _ "B")) values: ((literal) (#eq? _ "2"))))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestValueSpec(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
func main() {
	//asq_start
	var x, y int = 1, 2
	//asq_end
}`
	var err error
	var file *ast.File
	file, err = parser.ParseFile(fset, "", src, parser.ParseComments)
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
	var node Node
	node = BuildAsqNode(valueSpec, newPassOne(file))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(value_spec names: ((identifier) (#eq? _ "x") (identifier) (#eq? _ "y")) type: (identifier) (#eq? _ "int") values: ((literal) (#eq? _ "1") (literal) (#eq? _ "2")))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestChanType(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
func main() {
	//asq_start
	var ch chan int
	//asq_end
}`
	var err error
	var file *ast.File
	file, err = parser.ParseFile(fset, "", src, parser.ParseComments)
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
	var node Node
	node = BuildAsqNode(chanType, newPassOne(file))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(channel_type value: (identifier) (#eq? _ "int"))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestCompositeLit(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
func main() {
	//asq_start
	x := []int{1, 2}
	//asq_end
}`
	var err error
	var file *ast.File
	file, err = parser.ParseFile(fset, "", src, parser.ParseComments)
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
	var node Node
	node = BuildAsqNode(compositeLit, newPassOne(file))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(composite_literal type: (array_type element: (identifier) (#eq? _ "int")) elements: ((literal) (#eq? _ "1") (literal) (#eq? _ "2")))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestFuncDecl(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
//asq_start
func Add(x, y int) int {
	return x + y
}
//asq_end`
	var err error
	var file *ast.File
	file, err = parser.ParseFile(fset, "", src, parser.ParseComments)
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
	var node Node
	node = BuildAsqNode(funcDecl, newPassOne(file))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(function_declaration name: (identifier) (#eq? _ "Add") parameters: (field_list (field_declaration names: ((identifier) (#eq? _ "x") (identifier) (#eq? _ "y")) type: (identifier) (#eq? _ "int"))) results: (field_list (field_declaration type: (identifier) (#eq? _ "int"))) body: (block (return_statement (expression_list (binary_expression left: (identifier) (#eq? _ "x") operator: "+" right: (identifier) (#eq? _ "y"))))))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestFuncLit(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
func main() {
	//asq_start
	f := func(x int) int { return x * 2 }
	//asq_end
}`
	var err error
	var file *ast.File
	file, err = parser.ParseFile(fset, "", src, parser.ParseComments)
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
	var node Node
	node = BuildAsqNode(funcLit, newPassOne(file))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(function_literal type: (function_type parameters: (field_list (field_declaration names: ((identifier) (#eq? _ "x")) type: (identifier) (#eq? _ "int"))) results: (field_list (field_declaration type: (identifier) (#eq? _ "int")))))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestPackage(t *testing.T) {
	fset := token.NewFileSet()
	src := `//asq_start
package test
//asq_end`
	var err error
	var file *ast.File
	file, err = parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	var buf bytes.Buffer
	var node Node
	node = BuildAsqNode(file, newPassOne(file))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(source_file package_name: (identifier) (#eq? _ "test"))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestTypeSpec(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
//asq_start
type Point struct {
	X, Y int
}
//asq_end`
	var err error
	var file *ast.File
	file, err = parser.ParseFile(fset, "", src, parser.ParseComments)
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
	var node Node
	node = BuildAsqNode(typeSpec, newPassOne(file))
	if err := node.WriteTreeSitterQuery(&buf); err != nil {
		t.Fatalf("WriteTreeSitterQuery failed: %v", err)
	}

	expected := `(type_spec name: (identifier) (#eq? _ "Point") type: (struct_type fields: (field_list (field_declaration names: ((identifier) (#eq? _ "X") (identifier) (#eq? _ "Y")) type: (identifier) (#eq? _ "int")))))`
	if got := buf.String(); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}
