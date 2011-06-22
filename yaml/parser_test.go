package yaml

import (
	"testing"
	"bytes"
	"fmt"

	"runtime/debug"
)

func TestNoop(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
		}
	}()

	/*
		buf := bytes.NewBufferString(`
	foo: fooval
	bar:
	  - blah: test
	    boo: hoo
	  - baz
	  - long

	    text text
	  - bazor
	  - bazes: lorem ipsum
	    ipsum: dolor sit amet
	  - sit: amet
	  - - blah
	    - blah
	  - -
	    - no
	`)
	*/
	buf := bytes.NewBufferString(`
key1: val1
key2: val2
key3:
  subkey1: subval1
  subkey2: subval2
key4: val4
nested1:
  nested2:
    nested3: text
`)
	node, err := Parse(buf)
	if err != nil {
		t.Errorf("parse: %s", err)
	}
	fmt.Println(node)
}
