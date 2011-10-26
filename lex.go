package bencode

import (
	"strings"
	"fmt"
	"strconv"
)

type token struct {
	typ tokenType
	val string
}

func (t *token) String() string {
	return fmt.Sprintf("[%s]%q", t.typ.String(), t.val)
}

type tokenType int

const (
	eofType = tokenType(iota)
	errorType
	stringType
	intType
	dictStartType
	dictEndType
	listStartType
	listEndType
)

var tokenNames = map[tokenType]string{
	eofType:       "EOF",
	errorType:     "error",
	stringType:    "string",
	intType:       "int",
	dictStartType: "dictStart",
	dictEndType:   "dictEnd",
	listStartType: "listStart",
	listEndType:   "listEnd",
}

func (t *tokenType) String() string {
	return tokenNames[*t]
}

type tokenStack []tokenType

func (ts *tokenStack) push(t tokenType) {
	*ts = append(tokenStack{t}, *ts...)
}
func (t *tokenStack) pop() tokenType {
	val := (*t)[0]
	*t = (*t)[1:]
	return val
}

const eof = -1

type stateFn func(*lexer) stateFn

type lexer struct {
	input      string
	state      stateFn
	pos        int
	start      int
	width      int
	items      chan token
	peekBuffer *token
	eofd       bool
	stack      tokenStack
}

func (l *lexer) next() (rune int) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	rune, l.width = int(l.input[l.pos]), 1 //screw utf8
	l.pos += l.width
	return rune
}

func (l *lexer) peek() int {
	rune := l.next()
	l.backup()
	return rune
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) emit(t tokenType) {
	if t == eofType || t == errorType {
		l.eofd = true
	}
	l.items <- token{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) acceptLength(n int) {
	for ; n > 0; n-- {
		l.next()
	}
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- token{errorType, fmt.Sprintf(format, args...)}
	return nil
}

func (l *lexer) nextToken() token {
	if l.eofd {
		return token{eofType, ""}
	}
	if l.peekBuffer != nil {
		tmp := *l.peekBuffer
		l.peekBuffer = nil
		return tmp
	}

	return <-l.items
}

func (l *lexer) peekToken() token {
	if l.eofd {
		return token{eofType, ""}
	}
	if l.peekBuffer != nil {
		return *l.peekBuffer
	}
	tmp := <-l.items
	l.peekBuffer = &tmp
	return tmp
}

func lex(input string) *lexer {
	l := &lexer{
		input: input,
		state: lexValue,
		items: make(chan token),
		stack: make(tokenStack, 0),
	}

	go func() {
		for l.state != nil {
			l.state = l.state(l)
		}
	}()

	return l
}

func lexValue(l *lexer) stateFn {
	n := l.next()
	switch n {
	case eof:
		l.emit(eofType)
		return nil
	case 'i':
		l.ignore()
		return lexNumber
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return lexString
	case 'l':
		l.emit(listStartType)
		l.stack.push(listEndType)
		return lexValue
	case 'd':
		l.emit(dictStartType)
		l.stack.push(dictEndType)
		return lexValue
	case 'e':
		l.emit(l.stack.pop())
		return lexValue
	}
	return nil
}

func lexString(l *lexer) stateFn {
	//grab the rest of the numbers
	l.acceptRun("0123456789")
	length := l.input[l.start:l.pos]
	l.ignore()
	if l.next() != ':' {
		return l.errorf("Invalid format. Got a string list not followed by ':'")
	}
	l.ignore()
	n, err := strconv.Atoi(length)
	if err != nil {
		return l.errorf("Error parsing string len: %s", err)
	}

	l.acceptLength(n)
	l.emit(stringType)

	return lexValue
}

func lexNumber(l *lexer) stateFn {
	l.acceptRun("0123456789")
	l.emit(intType)

	if l.next() != 'e' {
		return l.errorf("Malformed integer")
	}
	l.ignore()

	return lexValue
}
