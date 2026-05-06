package parse

import (
	"reflect"
	"testing"
)

const pythonSample = `"""Module docstring."""
import os

CONSTANT = 42

def hello(name):
    """Greet."""
    return f"hi {name}"

async def fetch(url):
    return url

@staticmethod
def utility():
    return 1

class Greeter:
    def greet(self, name):
        return name

@dataclass
class Point:
    x: int
    y: int
`

func TestParsePythonTopLevel(t *testing.T) {
	got, err := ParsePythonTopLevel([]byte(pythonSample))
	if err != nil {
		t.Fatalf("ParsePythonTopLevel: %v", err)
	}
	want := []Symbol{
		{Name: "hello", Qualified: "hello", Kind: KindFunction, StartLine: 6, EndLine: 8},
		{Name: "fetch", Qualified: "fetch", Kind: KindFunction, StartLine: 10, EndLine: 11},
		{Name: "utility", Qualified: "utility", Kind: KindFunction, StartLine: 13, EndLine: 15},
		{Name: "Greeter", Qualified: "Greeter", Kind: KindClass, StartLine: 17, EndLine: 19},
		{Name: "Point", Qualified: "Point", Kind: KindClass, StartLine: 21, EndLine: 24},
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
