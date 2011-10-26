package bencode

import "strconv"

type Consumer struct {
	l *lexer
}

func Consume(data string) *Consumer {
	return &Consumer{
		l: lex(data),
	}
}

func (c *Consumer) Next() interface{} {
	return nextValue(c.l)
}

func nextValue(l *lexer) interface{} {
	next := l.nextToken()
	switch next.typ {
	case intType:
		n, err := strconv.Atoi(next.val)
		if err != nil {
			return nil
		}
		return n
	case stringType:
		return next.val
	case listStartType:
		return consumeList(l)
	case dictStartType:
		return consumeDict(l)
	case eofType:
		return nil
	case errorType:
		return nil
	}

	return nil
}

func consumeDict(l *lexer) map[string]interface{} {
	ret := make(map[string]interface{})

	for {
		key := l.nextToken()
		switch key.typ {
		case dictEndType:
			return ret
		default:
			break
		}

		switch l.peekToken().typ {
		case eofType:
			break
		case errorType:
			break
		case dictEndType:
			break
		}

		ret[key.val] = nextValue(l)
	}

	return nil
}

func consumeList(l *lexer) []interface{} {
	ret := make([]interface{}, 0)
	for {
		next := l.peekToken()
		switch next.typ {
		case eofType:
			break
		case errorType:
			break
		case listEndType:
			//consume it
			l.nextToken()
			return ret
		}

		ret = append(ret, nextValue(l))
	}

	return nil
}