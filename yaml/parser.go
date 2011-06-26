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
	lb := &LineBuffer{
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

	fmt.Println("Parse: parsing node")

	node = parseNode(lb, 0, nil)
	return
}

type Line struct {
	lineno int
	indent int
	line   []byte
}

func (line *Line) String() string {
	return fmt.Sprintf("%2d: %s%s", line.indent,
		strings.Repeat(" ", 0*line.indent), string(line.line))
}

func (line *Line) BreakAt(i int) (end *Line) {
	end = new(Line)
	*end = *line
	line.line = line.line[:i]
	end.line = end.line[i:]
	return
}

type LineReader interface {
	Next(minIndent int) *Line
}

const (
	typUnknown = iota
	typSequence
	typMapping
	typScalar
)

var typNames = []string{
	"Unknown", "Sequence", "Mapping", "Scalar",
}

func parseNode(r LineReader, ind int, initial Node) (node Node) {
	first := true
	node = initial

	// read lines
	for {
		line := r.Next(ind)
		if line == nil {
			break
		}
		pfx := strings.Repeat(".", line.indent)

		if len(line.line) == 0 {
			continue
		}
		fmt.Printf("%s%#v (initial)\n", strings.Repeat("+", line.indent), node)
		fmt.Printf("%s%s\n", strings.Repeat("=", line.indent), string(line.line))

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
			end = bytes.TrimLeft(end, " \t")

			switch vtyp {
			case typScalar:
				//fmt.Printf("%s=Scalar: %q\n", pfx, end)
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

			/*
			inline := parseNode(r, line.indent+1, typUnknown)
			fmt.Printf("%sSUB: %#v\n", pfx, inline)
			switch vtyp {
			case typMapping:
				if inline == nil {
					inline = make(Map)
				}
				mapNode, ok := inline.(Map)
				if !ok {
					panic(fmt.Sprintf("type mismatch: %T + %T", mapNode, inline))
				}
				//fmt.Printf("%s=Mapping %T=%#v %v\n", pfx, inline, mapNode, ok)
			}
			*/
		}

		inlineValue(line.line)
		var prev Node

		// Nest inlines
		for len(types) > 0 {
			fmt.Printf("%sTYP: %v\n", pfx, types)
			fmt.Printf("%sVAL: %v\n", pfx, pieces)

			last := len(types)-1
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
					fmt.Printf("%sInline: %#v\n", pfx, current)
					break
				}

				child = parseNode(r, line.indent+1, prev)
				mapNode[piece] = child
				current = mapNode

				fmt.Printf("%sAssign %q: %#v\n", pfx, piece, current)
			}

			if last < 0 {
				last = 0
			}
			types = types[:last]
			pieces = pieces[:last]
			fmt.Printf("%sINL: %#v\n", pfx, current)
			prev = current
		}

		fmt.Printf("%sLIN: %#v\n", pfx, prev)
		node = prev
	}
	fmt.Printf("%sRET: %#v\n", strings.Repeat(":", ind), node)
	return
}

func getType(line []byte) (typ, split int) {
	if len(line) == 0 {
		return
	}
	if line[0] == '-' {
		typ = typSequence
		split = 1
	} else {
		for i := 0; i < len(line); i++ {
			switch ch := line[i]; ch {
			case ' ', '\t':
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

// LineReader implementations

type LineBuffer struct {
	*bufio.Reader
	readLines int
	pending *Line
}

func (lb *LineBuffer) Next(min int) (next *Line) {
	if lb.pending == nil {
		var (
			read []byte
			more bool
			err  os.Error
		)

		l := new(Line)
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
			case ' ', '\t':
				l.indent += 1
				continue
			default:
			}
			break
		}
		l.line = l.line[l.indent:]
		lb.pending = l
	}
	next = lb.pending
	if next.indent < min {
		return nil
	}
	lb.pending = nil
	return
}

type LineSlice []*Line

func (ls *LineSlice) Next(min int) (next *Line) {
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

func (ls *LineSlice) Push(line *Line) {
	*ls = append(*ls, line)
}
