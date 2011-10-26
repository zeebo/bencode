package bencode

import (
	"strconv"
	"os"
)

func nextValue(l *lexer) (interface{}, os.Error) {
	switch next := l.nextToken(); next.typ {
	case intType:
		n, err := strconv.Atoi(next.val)
		if err != nil {
			return nil, err
		}
		return n, nil
	case stringType:
		return next.val, nil
	case listStartType:
		return consumeList(l)
	case dictStartType:
		return consumeDict(l)
	case eofType:
		return nil, os.EOF
	case errorType:
		return nil, next
	}

	return nil, os.NewError("Unknown type")
}

func consumeDict(l *lexer) (map[string]interface{}, os.Error) {
	ret := make(map[string]interface{})

	for {
		key := l.nextToken()
		switch key.typ {
		case dictEndType:
			return ret, nil
		case eofType:
			return nil, os.NewError("Unexpected EOF")
		case errorType:
			return nil, key
		}

		switch l.peekToken().typ {
		case eofType:
			return nil, os.NewError("Unexpected EOF")
		case errorType:
			return nil, l.nextToken() //consume the token
		case dictEndType:
			return nil, os.NewError("Unexpected Dict End")
		}

		val, err := nextValue(l)
		if err != nil {
			return nil, err
		}
		ret[key.val] = val
	}

	panic("unreachable")
}

func consumeList(l *lexer) ([]interface{}, os.Error) {
	ret := make([]interface{}, 0)
	for {
		switch next := l.peekToken(); next.typ {
		case eofType:
			return nil, os.NewError("Unexpected EOF")
		case errorType:
			return nil, next
		case listEndType:
			//consume it
			l.nextToken()
			return ret, nil
		}

		val, err := nextValue(l)
		if err != nil {
			return nil, err
		}
		ret = append(ret, val)
	}

	panic("unreachable")
}