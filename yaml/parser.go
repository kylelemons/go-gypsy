package yaml

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

func Parse(r io.Reader) (node *Node, err os.Error) {
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case os.Error: err = r
			case string:   err = os.NewError("yaml: " + r)
			default:       err = fmt.Errorf("%v", r)
			}
		}
	}()
	p := &parser{
		Reader: bufio.NewReader(r),
		Stop:   make(chan bool),
		Tokens: make(chan token, 10),
	}
	go p.tokenize()
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
	Type   int
	String string
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
)

var tokenNames = map[int]string{
	tokIndent: "INDENT",
	tokSpace: "SPACE",
	tokLabel:  "LABEL",
	tokString: "STRING",
	tokNewline: "EOL",
	tokColon: "COLON",
	tokDash: "DASH",
	tokListOpen: "LIST-OPEN",
	tokListClose: "LIST-CLOSE",
	tokMapOpen: "MAP-OPEN",
	tokMapClose: "MAP-CLOSE",
	tokComma: "COMMA",
}

func (p *parser) next() (tok token) {
	if p.Last != nil {
		tok, p.Last = *p.Last, nil
		return
	}
	tok = <-p.Tokens
	return
}

func (p *parser) backup(tok token) {
	p.Last = &tok
}

func (p *parser) tokenize() {
	defer close(p.Tokens)

	var (
		line, part []byte
		err os.Error
		more, initialSpace bool
		typ int
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
			p.Tokens <- token{typ, string(line[i:i+1])}
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
			p.Tokens <- token{typ, string(line[i:i+1])}
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
}

func (p *parser) pop() {
	if n := len(p.Prefix); n > 0 {
		p.Prefix = p.Stack[n-1]
		p.Stack = p.Stack[:n-1]
	}
}

func (p *parser) parseNode() *Node {
	for tok := range p.Tokens {
		t := tokenNames[tok.Type]
		s := tok.String
		fmt.Printf("(%s)%s", t, s)
	}

	

	return nil
}
