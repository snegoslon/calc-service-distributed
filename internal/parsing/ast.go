package parsing

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type Node struct {
	IsLeaf        bool
	Value         float64
	Operator      string
	Left, Right   *Node
	TaskScheduled bool
}

func ParseExpression(expression string) (*Node, error) {
	expr := strings.ReplaceAll(expression, " ", "")
	if expr == "" {
		return nil, fmt.Errorf("empty expression")
	}
	p := &parsingState{input: expr, pos: 0}
	node, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if p.pos < len(p.input) {
		return nil, fmt.Errorf("unexpected token at position %d", p.pos)
	}
	return node, nil
}

type parsingState struct {
	input string
	pos   int
}

func (p *parsingState) peek() rune {
	if p.pos < len(p.input) {
		return rune(p.input[p.pos])
	}
	return 0
}

func (p *parsingState) get() rune {
	ch := p.peek()
	p.pos++
	return ch
}

func (p *parsingState) parseExpression() (*Node, error) {
	node, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for {
		ch := p.peek()
		if ch == '+' || ch == '-' {
			op := string(p.get())
			right, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			node = &Node{
				IsLeaf:   false,
				Operator: op,
				Left:     node,
				Right:    right,
			}
		} else {
			break
		}
	}
	return node, nil
}

func (p *parsingState) parseTerm() (*Node, error) {
	node, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for {
		ch := p.peek()
		if ch == '*' || ch == '/' {
			op := string(p.get())
			right, err := p.parseFactor()
			if err != nil {
				return nil, err
			}
			node = &Node{
				IsLeaf:   false,
				Operator: op,
				Left:     node,
				Right:    right,
			}
		} else {
			break
		}
	}
	return node, nil
}

func (p *parsingState) parseFactor() (*Node, error) {
	ch := p.peek()
	if ch == '(' {
		p.get()
		node, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.peek() != ')' {
			return nil, fmt.Errorf("missing closing parenthesis")
		}
		p.get()
		return node, nil
	}
	start := p.pos
	if ch == '+' {
		if p.pos > 0 && p.input[p.pos-1] != '(' {
			return nil, fmt.Errorf("unexpected unary plus at position %d", p.pos)
		}
		p.get()
	} else if ch == '-' {
		p.get()
	}
	for {
		ch = p.peek()
		if unicode.IsDigit(ch) || ch == '.' {
			p.get()
		} else {
			break
		}
	}
	token := p.input[start:p.pos]
	if token == "" {
		return nil, fmt.Errorf("expected number at position %d", start)
	}
	value, err := strconv.ParseFloat(token, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid number %s", token)
	}
	return &Node{
		IsLeaf: true,
		Value:  value,
	}, nil
}