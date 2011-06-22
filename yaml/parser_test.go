package yaml

import (
	"testing"
	"bytes"
)

func TestNoop(t *testing.T) {
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
	Parse(buf)
}
