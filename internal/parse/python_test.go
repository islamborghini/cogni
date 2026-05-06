package parse

import (
	"embed"
	"reflect"
	"testing"
)

//go:embed testdata/*.py
var fixtures embed.FS

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := fixtures.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("load fixture %s: %v", name, err)
	}
	return data
}

func TestParsePythonTopLevel(t *testing.T) {
	got, err := ParsePythonTopLevel(loadFixture(t, "simple.py"))
	if err != nil {
		t.Fatalf("ParsePythonTopLevel: %v", err)
	}
	want := []Symbol{
		{Name: "hello", Qualified: "hello", Kind: KindFunction, StartLine: 6, EndLine: 8, Signature: "def hello(name)", Docstring: "Greet."},
		{Name: "fetch", Qualified: "fetch", Kind: KindFunction, StartLine: 10, EndLine: 11, Signature: "async def fetch(url)"},
		{Name: "utility", Qualified: "utility", Kind: KindFunction, StartLine: 13, EndLine: 15, Signature: "def utility()"},
		{Name: "Greeter", Qualified: "Greeter", Kind: KindClass, StartLine: 17, EndLine: 19},
		{Name: "Point", Qualified: "Point", Kind: KindClass, StartLine: 21, EndLine: 24},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("symbol mismatch\n got:  %+v\n want: %+v", got, want)
	}
}

func TestParsePython_Simple(t *testing.T) {
	got, err := ParsePython(loadFixture(t, "simple.py"))
	if err != nil {
		t.Fatalf("ParsePython: %v", err)
	}
	want := []Symbol{
		{Name: "CONSTANT", Qualified: "CONSTANT", Kind: KindConstant, StartLine: 4, EndLine: 4},
		{Name: "hello", Qualified: "hello", Kind: KindFunction, StartLine: 6, EndLine: 8, Signature: "def hello(name)", Docstring: "Greet."},
		{Name: "fetch", Qualified: "fetch", Kind: KindFunction, StartLine: 10, EndLine: 11, Signature: "async def fetch(url)"},
		{Name: "utility", Qualified: "utility", Kind: KindFunction, StartLine: 13, EndLine: 15, Signature: "def utility()"},
		{Name: "Greeter", Qualified: "Greeter", Kind: KindClass, StartLine: 17, EndLine: 19},
		{Name: "greet", Qualified: "Greeter.greet", Kind: KindMethod, StartLine: 18, EndLine: 19, Signature: "def greet(self, name)"},
		{Name: "Point", Qualified: "Point", Kind: KindClass, StartLine: 21, EndLine: 24},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("symbol mismatch\n got:  %+v\n want: %+v", got, want)
	}
}

func TestParsePython_Nested(t *testing.T) {
	got, err := ParsePython(loadFixture(t, "nested.py"))
	if err != nil {
		t.Fatalf("ParsePython: %v", err)
	}
	want := []Symbol{
		{Name: "Outer", Qualified: "Outer", Kind: KindClass, StartLine: 3, EndLine: 16, Docstring: "Outer class."},
		{Name: "Inner", Qualified: "Outer.Inner", Kind: KindClass, StartLine: 6, EndLine: 10, Docstring: "Inner class."},
		{Name: "ping", Qualified: "Outer.Inner.ping", Kind: KindMethod, StartLine: 9, EndLine: 10, Signature: "def ping(self)"},
		{Name: "outer_method", Qualified: "Outer.outer_method", Kind: KindMethod, StartLine: 12, EndLine: 16, Signature: "def outer_method(self, x)", Docstring: "Outer method."},
		{Name: "StandAlone", Qualified: "StandAlone", Kind: KindClass, StartLine: 18, EndLine: 19},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("symbol mismatch\n got:  %+v\n want: %+v", got, want)
	}
}

func TestParsePythonTopLevel_Empty(t *testing.T) {
	got, err := ParsePythonTopLevel([]byte("# just a comment\n"))
	if err != nil {
		t.Fatalf("ParsePythonTopLevel: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no symbols, got %+v", got)
	}
}
