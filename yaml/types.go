package yaml

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type Node interface {
	Map(string) (Node,bool)
	List() []Node
	String() string
	write(io.Writer,int,int)
}

type Map map[string]Node
func (node Map) Map(key string) (val Node, ok bool) {
	val, ok = node[key]
	return
}
func (node Map) List() []Node {
	return nil
}
func (node Map) String() string {
	out := bytes.NewBuffer(nil)
	node.write(out, 0, 0)
	return out.String()
}
func (node Map) write(out io.Writer, firstind, nextind int) {
	indent := bytes.Repeat([]byte{' '}, nextind)
	ind := firstind

	width := 0
	scalarkeys := []string{}
	objectkeys := []string{}
	for key, value := range node {
		if _, ok := value.(Scalar); ok {
			if swid := len(key); swid > width {
				width = swid
			}
			scalarkeys = append(scalarkeys, key)
			continue
		}
		objectkeys = append(objectkeys, key)
	}
	sort.SortStrings(scalarkeys)
	sort.SortStrings(objectkeys)

	for _, key := range scalarkeys {
		value := node[key].(Scalar)
		out.Write(indent[:ind])
		fmt.Fprintf(out, "%-*s %s\n", width+1, key+":", string(value))
		ind = nextind
	}
	for _, key := range objectkeys {
		out.Write(indent[:ind])
		fmt.Fprintf(out, "%s:\n", key)
		ind = nextind
		node[key].write(out, ind+2, ind+2)
	}
}

type List []Node
func (node List) Map(key string) (Node,bool) { return nil, false }
func (node List) List() []Node {
	return node
}
func (node List) String() string {
	out := bytes.NewBuffer(nil)
	node.write(out, 0, 0)
	return out.String()
}
func (node List) write(out io.Writer, firstind, nextind int) {
	indent := bytes.Repeat([]byte{' '}, nextind)
	ind := firstind

	for _, value := range node {
		out.Write(indent[:ind])
		fmt.Fprint(out, "- ")
		ind = nextind
		value.write(out, 0, ind+2)
	}
}

type Scalar string
func (node Scalar) Map(_ string) (Node,bool) { return nil, false }
func (node Scalar) List() ([]Node) { return nil }
func (node Scalar) String() string { return string(node) }
func (node Scalar) write(out io.Writer, ind, _ int) {
	fmt.Fprintf(out, "%s%s\n", strings.Repeat(" ", ind), string(node))
}

func Scan(format string, args ...interface{}) (n int, err os.Error) {
	return
}
