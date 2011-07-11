package yaml

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"bytes"
)

func Parse(r io.Reader) (node Node, err os.Error) {
	lb := &lineBuffer{
		Reader: bufio.NewReader(r),
	}

	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case os.Error:
				err = r
			case string:
				err = os.NewError(r)
			default:
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	node = parseNode(lb, 0, nil)
	return
}

// Supporting types and constants

const (
	typUnknown = iota
	typSequence
	typMapping
	typScalar
)

var typNames = []string{
	"Unknown", "Sequence", "Mapping", "Scalar",
}

type lineReader interface {
	Next(minIndent int) *indentedLine
}

type indentedLine struct {
	lineno int
	indent int
	line   []byte
}

func (line *indentedLine) String() string {
	return fmt.Sprintf("%2d: %s%s", line.indent,
		strings.Repeat(" ", 0*line.indent), string(line.line))
}

func parseNode(r lineReader, ind int, initial Node) (node Node) {
	first := true
	node = initial

	// read lines
	for {
		line := r.Next(ind)
		if line == nil {
			break
		}

		if len(line.line) == 0 {
			continue
		}

		if first {
			ind = line.indent
			first = false
		}

		types := []int{}
		pieces := []string{}

		var inlineValue func([]byte)
		inlineValue = func(partial []byte) {
			// TODO(kevlar): This can be a for loop now
			vtyp, brk := getType(partial)
			begin, end := partial[:brk], partial[brk:]

			if vtyp == typMapping {
				end = end[1:]
			}
			end = bytes.TrimLeft(end, " ")

			switch vtyp {
			case typScalar:
				types = append(types, typScalar)
				pieces = append(pieces, string(end))
				return
			case typMapping:
				types = append(types, typMapping)
				pieces = append(pieces, string(begin))
				inlineValue(end)
			case typSequence:
				types = append(types, typSequence)
				pieces = append(pieces, "-")
				inlineValue(end)
			}
		}

		inlineValue(line.line)
		var prev Node

		// Nest inlines
		for len(types) > 0 {
			last := len(types) - 1
			typ, piece := types[last], pieces[last]

			var current Node
			if last == 0 {
				current = node
			}
			//child := parseNode(r, line.indent+1, typUnknown) // TODO allow scalar only

			// Add to current node
			switch typ {
			case typScalar: // last will be == nil
				if _, ok := current.(Scalar); current != nil && !ok {
					panic("cannot append scalar to non-scalar node")
				}
				if current != nil {
					current = Scalar(piece) + " " + current.(Scalar)
					break
				}
				current = Scalar(piece)
			case typMapping:
				var mapNode Map
				var ok bool
				var child Node

				// Get the current map, if there is one
				if mapNode, ok = current.(Map); current != nil && !ok {
					_ = current.(Map) // panic
				} else if current == nil {
					mapNode = make(Map)
				}

				if _, inlineMap := prev.(Scalar); inlineMap && last > 0 {
					current = Map{
						piece: prev,
					}
					break
				}

				child = parseNode(r, line.indent+1, prev)
				mapNode[piece] = child
				current = mapNode

			case typSequence:
				var listNode List
				var ok bool
				var child Node

				// Get the current list, if there is one
				if listNode, ok = current.(List); current != nil && !ok {
					_ = current.(List) // panic
				} else if current == nil {
					listNode = make(List, 0)
				}

				if _, inlineList := prev.(Scalar); inlineList && last > 0 {
					current = List{
						prev,
					}
					break
				}

				child = parseNode(r, line.indent+1, prev)
				listNode = append(listNode, child)
				current = listNode

			}

			if last < 0 {
				last = 0
			}
			types = types[:last]
			pieces = pieces[:last]
			prev = current
		}

		node = prev
	}
	return
}

func getType(line []byte) (typ, split int) {
	if len(line) == 0 {
		return
	}
	if line[0] == '-' {
		typ = typSequence
		split = 1
		return
	} else {
		for i := 0; i < len(line); i++ {
			switch ch := line[i]; ch {
			case ' ':
				typ = typScalar
			case ':':
				typ = typMapping
				split = i
			default:
				continue
			}
			return
		}
	}
	typ = typScalar
	return
}

// lineReader implementations

type lineBuffer struct {
	*bufio.Reader
	readLines int
	pending   *indentedLine
}

func (lb *lineBuffer) Next(min int) (next *indentedLine) {
	if lb.pending == nil {
		var (
			read []byte
			more bool
			err  os.Error
		)

		l := new(indentedLine)
		l.lineno = lb.readLines
		more = true
		for more {
			read, more, err = lb.ReadLine()
			if err != nil {
				if err == os.EOF {
					return nil
				}
				panic(err)
			}
			l.line = append(l.line, read...)
		}
		lb.readLines++

		for _, ch := range l.line {
			switch ch {
			case ' ':
				l.indent += 1
				continue
			default:
			}
			break
		}
		l.line = l.line[l.indent:]

		// Ignore blank lines and comments.
		if len(l.line) == 0 || l.line[0] == '#' {
			return lb.Next(min)
		}

		lb.pending = l
	}
	next = lb.pending
	if next.indent < min {
		return nil
	}
	lb.pending = nil
	return
}

type lineSlice []*indentedLine

func (ls *lineSlice) Next(min int) (next *indentedLine) {
	if len(*ls) == 0 {
		return nil
	}
	next = (*ls)[0]
	if next.indent < min {
		return nil
	}
	*ls = (*ls)[1:]
	return
}

func (ls *lineSlice) Push(line *indentedLine) {
	*ls = append(*ls, line)
}
