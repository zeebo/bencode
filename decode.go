package bencode

import (
	"os"
	"strconv"
)

func decodeInto(l *lexer, val interface{}) os.Error {
	switch next := l.peekToken(); next.typ {
	case eofType:
		return os.EOF
	case errorType:
		return next
	case intType:
		return decodeInt(l, val)
	case stringType:
		return decodeString(l, val)
	case listStartType:
		return decodeList(l, val)
	case dictStartType:
		return decodeDict(l, val)
	}

	panic("unknown token")
}

func decodeInt(l *lexer, val interface{}) os.Error {
	token := l.nextToken()

	n, err := strconv.Atoi64(token.val)
	if err != nil {
		return err
	}

	//store n in val somehow
	val = n

	return nil
}

func decodeString(l *lexer, val interface{}) os.Error {
	token := l.nextToken()

	//store token.val in val somehow
	val = token.val

	return nil
}

func decodeList(l *lexer, val interface{}) os.Error {
	var nextVal interface{}
	for {
		switch next := l.peekToken(); next.typ {
		case listEndType:
			return nil
		case eofType:
			return os.NewError("Unexpected EOF")
		case errorType:
			return l.nextToken() //consume the error
		}

		//create an empty thing, decodeInto(l, thing), append to val
		if err := decodeInto(l, &nextVal); err != nil {
			return err
		}

		if err := appendIntoList(val, nextVal); err != nil {
			return err
		}
	}

	panic("unreachable")
}

func appendIntoList(list, val interface{}) os.Error {
	return nil
}

func decodeDict(l *lexer, val interface{}) os.Error {
	var nextVal interface{}
	for {
		key := l.nextToken()
		switch key.typ {
		case dictEndType:
			return nil
		case eofType:
			return os.NewError("Unexpected EOF")
		case errorType:
			return key
		}

		switch l.peekToken().typ {
		case eofType:
			return os.NewError("Unexpected EOF")
		case dictEndType:
			return os.NewError("Unexpected Dict End")
		case errorType:
			return l.nextToken() //consume the error
		}

		//create an empty thing, decodeInto(l, thing), store into appropriate part of val
		if err := decodeInto(l, &nextVal); err != nil {
			return err
		}

		if err := storeIntoDict(val, key.val, nextVal); err != nil {
			return err
		}
	}

	panic("unreachable")
}

func storeIntoDict(dict interface{}, key string, val interface{}) os.Error {
	return nil
}

