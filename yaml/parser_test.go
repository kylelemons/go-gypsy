package yaml

import (
	"testing"
	"bytes"
	"fmt"

	"runtime/debug"
)

var parseTests = []struct {
	Input  string
	Output string
}{
	{
		Input:
			"key1: val1\n",
		Output:
			"key1: val1\n",
	},
	{
		Input:
			"key: nest: val\n",
		Output:
			"key:\n"+
			"  nest: val\n",
	},
	{
		Input:
			"a: b: c: d\n"+
			"      e: f\n"+
			"   g: h: i\n"+
			"      j: k\n"+
			"   l: m\n"+
			"n: o\n"+
			"",
		Output:
			"n: o\n"+
			"a:\n"+
			"  l: m\n"+
			"  b:\n"+
			"    c: d\n"+
			"    e: f\n"+
			"  g:\n"+
			"    h: i\n"+
			"    j: k\n"+
			"",
	},
}

func TestParse(t *testing.T) {
	/*
		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
			}
		}()
	*/
	_ = debug.PrintStack

	for idx, test := range parseTests {
		buf := bytes.NewBufferString(test.Input)
		node, err := Parse(buf)
		if err != nil {
			t.Errorf("parse: %s", err)
		}
		buf.Truncate(0)
		fmt.Fprintf(buf, "%s", node)
		if got, want := buf.String(), test.Output; got != want {
			t.Errorf("---%d---", idx)
			t.Errorf("got: %q:\n%s", got, got)
			t.Errorf("want: %q:\n%s", want, want)
		}
	}
}

var lineBreakTests = []struct {
	Line    string
	BreakAt int
	First   string
	Last    string
}{
	{"- blah: test", 2, "- ", "blah: test"},
	{"  - - test", 2, "- ", "- test"},
}

func TestLineBreak(t *testing.T) {
	for idx, test := range lineBreakTests {
		var first, last *Line
		var i int
		for ; i < len(test.Line); i++ {
			if ch := test.Line[i]; ch != ' ' {
				first = &Line{
					lineno: idx,
					indent: i,
					line:   []byte(test.Line[i:]),
				}
				break
			}
		}
		last = first.BreakAt(test.BreakAt)
		if got, want := string(first.line), test.First; got != want {
			t.Errorf("%d. first = %q, want %q", got, want)
		}
		if got, want := string(last.line), test.Last; got != want {
			t.Errorf("%d. last = %q, want %q", got, want)
		}
	}
}
