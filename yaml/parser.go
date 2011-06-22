package yaml

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

func Parse(r io.Reader) (node Node, err os.Error) {
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case os.Error:
				err = r
			case string:
				err = os.NewError("yaml: " + r)
			default:
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	p := &parser{
		Reader: bufio.NewReader(r),
		Stop:   make(chan bool),
		Tokens: make(chan token, 10),
	}
	go p.tokenize()
	fmt.Println("Parse: parsing node")
	node = p.parseNode()
	return
}

type parser struct {
	*bufio.Reader

	Prefix string
	Stack  []string
	Stop   chan bool
	Tokens chan token
	Last   *token
}

type token struct {
	Type int
	Str  string
}

func (tok token) String() string {
	return fmt.Sprintf("[%s:%q]", tokenNames[tok.Type], tok.Str)
}

const (
	tokIndent = iota
	tokSpace
	tokLabel
	tokString
	tokNewline
	tokColon
	tokDash
	tokListOpen
	tokListClose
	tokMapOpen
	tokMapClose
	tokComma
	tokEOF
)

var tokenNames = map[int]string{
	tokIndent:    "INDENT",
	tokSpace:     "SPACE",
	tokLabel:     "LABEL",
	tokString:    "STRING",
	tokNewline:   "EOL",
	tokColon:     "COLON",
	tokDash:      "DASH",
	tokListOpen:  "LIST-OPEN",
	tokListClose: "LIST-CLOSE",
	tokMapOpen:   "MAP-OPEN",
	tokMapClose:  "MAP-CLOSE",
	tokComma:     "COMMA",
	tokEOF:       "EOF",
}

func (p *parser) next() (tok token) {
	if p.Last != nil {
		tok, p.Last = *p.Last, nil
		return
	}
	var open bool
	tok, open = <-p.Tokens
	if !open {
		tok = token{tokEOF, ""}
	}
	return
}

func (p *parser) backup(tok token) {
	p.Last = &tok
}

func (p *parser) tokenize() {
	defer close(p.Tokens)

	var (
		line, part         []byte
		err                os.Error
		more, initialSpace bool
		typ                int
	)

lineLoop:
	for {
		select {
		case <-p.Stop:
			break
		default:
		}

		line, more, err = p.ReadLine()
		for more {
			var suffix []byte
			suffix, more, err = p.ReadLine()
			line = append(line, suffix...)
		}

		switch err {
		case os.EOF:
			break lineLoop
		default:
			panic(err)
		case nil:
		}

		initialSpace = true

	parseObject:
		// strip indent
		part = bytes.TrimLeft(line, " \t")

		if len(part) == 0 {
			p.Tokens <- token{tokNewline, "\n"}
			continue
		}
		if len(part) < len(line) {
			n, m := len(line), len(part)
			typ := tokSpace
			if initialSpace {
				typ = tokIndent
			}
			p.Tokens <- token{typ, string(line[:n-m])}
		}
		line = part

		initialSpace = false

		if line[0] == '-' {
			p.Tokens <- token{tokDash, "-"}
			line = line[1:]
			goto parseObject
		}

		for i := 0; i < len(line); i++ {
			switch line[i] {
			case ' ', '\t':
				typ = tokString
			case '[':
				typ = tokListOpen
			case '{':
				typ = tokMapOpen
			case ':':
				typ = tokColon
			default:
				continue
			}
			// if it's a string, consume the rest of the line
			if typ == tokString {
				p.Tokens <- token{tokString, string(line)}
				line = line[:0]
				break
			}
			if i > 0 {
				p.Tokens <- token{tokLabel, string(line[:i])}
			}
			p.Tokens <- token{typ, string(line[i : i+1])}
			line = line[i+1:]
			if typ == tokColon {
				goto parseObject
			}
			break
		}

		for i := 0; i < len(line); i++ {
			switch line[i] {
			case ' ', '\t':
				typ = tokString
			case '[':
				typ = tokListOpen
			case '{':
				typ = tokMapOpen
			case ']':
				typ = tokListClose
			case '}':
				typ = tokMapClose
			case ',':
				typ = tokComma
			case ':':
				typ = tokColon
			default:
				continue
			}
			// if it's a string, consume the rest of the line
			if typ == tokString {
				p.Tokens <- token{tokString, string(line)}
				line = line[:0]
				break
			}
			if i > 0 {
				p.Tokens <- token{tokString, string(line[:i])}
			}
			p.Tokens <- token{typ, string(line[i : i+1])}
			line = line[i+1:]
			i = 0
		}

		if len(line) > 0 {
			p.Tokens <- token{tokString, string(line)}
		}
		p.Tokens <- token{tokNewline, "\n"}
	}
}

func (p *parser) push() {
	p.Stack = append(p.Stack, p.Prefix)
	fmt.Printf("push: pushing prefix: %q %#v\n", p.Prefix, p.Stack)
}

func (p *parser) pop() {
	if n := len(p.Stack); n > 0 {
		p.Prefix = p.Stack[n-1]
		p.Stack = p.Stack[:n-1]
	}
	fmt.Printf("pop: new prefix: %q %#v\n", p.Prefix, p.Stack)
}

func (p *parser) parseNode() Node {
	/*
		for tok := range p.Tokens {
			t := tokenNames[tok.Type]
			s := tok.Str
			fmt.Printf("(%s)%s", t, s)
		}
	*/

	tok := p.next()

	if tok.Type == tokString {
		p.backup(tok)
		fmt.Println("parse: got string")
		return p.parseScalar()
	}

	gotIndent, wantIndent := "", p.Prefix
	for {
		fmt.Println("parse:", tok)
		switch tok.Type {
		case tokNewline:
			fmt.Println("parse: skip newline")
			tok = p.next()
			continue
		case tokIndent:
			fmt.Println("parse: skip indent")
			gotIndent += tok.Str
			tok = p.next()
			continue
		}
		break
	}

	// If we don't have sufficient indentation, this is a nil object
	if !strings.HasPrefix(gotIndent, wantIndent) {
		fmt.Println("parse: insufficient indent")
		return nil
	}
	p.push()
	p.Prefix = gotIndent
	defer p.pop()

	switch tok.Type {
	case tokLabel:
		p.backup(tok)
		fmt.Println("parse: got label")
		return p.parseMapping()
	default:
		fmt.Println("parse: unexpected token:", tok)
	}

	return nil
}

// prereqs:
//  - this line must have its indentation verified
//  - the label should be the first token in the stream
func (p *parser) parseMapping() (mapping Map) {
	for {
		tok := p.next()
		fmt.Println("map:", tok)

		// TODO(kevlar): factor into a function
		gotIndent, wantIndent := "", p.Prefix
		if tok.Type == tokIndent {
			fmt.Println("map: found indent")
			gotIndent = tok.Str
			tok = p.next()
		}

		switch len(mapping) {
		case 0:
			mapping = make(Map)
			fmt.Println("map: first keyval")
		case 1:
			// Set prefix based on first non-inline key
			p.Prefix = gotIndent
			fmt.Println("map: second keyval")
		default:
			fmt.Printf("map: indent: got %q, want %q\n", gotIndent, wantIndent)
			if gotIndent != wantIndent {
				p.backup(tok)
				fmt.Println("map: insufficient indent", tok)
				return
			}
			fmt.Println("map: subsequent keyval")
		}

		var key string
		var val Node

		switch tok.Type {
		case tokEOF:
			fmt.Println("map: EOF")
			return
		case tokLabel:
			key = tok.Str
			fmt.Println("map: key:", key)
		default:
			fmt.Println("map: want LABEL, got", tokenNames[tok.Type])
			return nil
		}

		if tok := p.next(); tok.Type != tokColon {
			fmt.Println("map: want COLON, got", tokenNames[tok.Type])
			return nil
		}

		for {
			if tok := p.next(); tok.Type != tokSpace {
				p.backup(tok)
				break
			}
		}

		val = p.parseNode()

		if val == nil {
			fmt.Printf("map: done %#v\n", mapping)
			break
		}
		mapping[key] = val
		fmt.Printf("map: added %#v\n", mapping)
	}
	return
}

// prereq:
// - next token should be string
func (p *parser) parseScalar() Scalar {
	str := ""
	for {
		tok := p.next()
		fmt.Println("scalar:", tok)
		switch tok.Type {
		case tokString:
			if len(str) > 0 {
				str += "\n"
			}
			str += tok.Str
			continue
		case tokNewline:
			continue
		// TODO(kevlar) multiline strings?
		/*
			case tokIndent:
				// TODO
				continue
		*/
		default:
			p.backup(tok)
		}
		break
	}
	fmt.Printf("scalar: %q\n", str)
	return Scalar(str)
}
