package bencode

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strconv"
)

var (
	reflectByteSliceType = reflect.TypeOf([]byte(nil))
	reflectStringType    = reflect.TypeOf("")
)

//A Decoder reads and decodes bencoded data from an input stream.
type Decoder struct {
	c *chunker
}

//NewDecoder returns a new decoder that reads from r
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{newChunker(r)}
}

//Decode reads the bencoded value from its input and stores it in the value pointed to by val.
//Decode allocates maps/slices as necessary with the following additional rules:
//To decode a bencoded value into a nil interface value, the type stored in the interface value is one of:
//	[u]int[8,16,32,64] for bencoded integers
//	string for bencoded strings
//	[]interface{} for bencoded lists
//	map[string]interface{} for bencoded dicts
func (d *Decoder) Decode(val interface{}) error {
	next, err := d.c.nextValue()
	if err != nil {
		return err
	}

	l := lex(next)

	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("Unwritable type passed into decode")
	}

	return decodeInto(l, rv)
}

//DecodeString reads the data in the string and stores it into the value pointed to by val.Errorf
//Read the docs for Decode for more information.
func DecodeString(in string, val interface{}) error {
	buf := bytes.NewBufferString(in)
	d := NewDecoder(buf)
	return d.Decode(val)
}

func indirect(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		if v.Kind() == reflect.Interface && !v.IsNil() {
			v = v.Elem()
			continue
		}
		if v.Kind() != reflect.Ptr {
			break
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

func decodeInto(l *lexer, val reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	var next token
	switch next = l.peekToken(); next.typ {
	case eofType:
		return io.EOF
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

	panic(fmt.Errorf("Unknown token: %s", next))
}

func decodeInt(l *lexer, val reflect.Value) error {
	token := l.nextToken()
	v := indirect(val)

	switch v.Kind() {
	default:
		return fmt.Errorf("Cannot store int64 into %s", v.Type())
	case reflect.Interface:
		n, err := strconv.ParseInt(token.val, 10, 64)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(n))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(token.val, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(token.val, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(n)
	}

	return nil
}

func decodeString(l *lexer, val reflect.Value) error {
	token := l.nextToken()
	v := indirect(val)

	switch v.Kind() {
	default:
		return fmt.Errorf("Cannot store string into %s", v.Type())
	case reflect.Slice:
		if v.Type() != reflectByteSliceType {
			return fmt.Errorf("Cannot store string into %s", v.Type())
		}
		v.Set(reflect.ValueOf([]byte(token.val)))
	case reflect.String:
		v.SetString(string(token.val))
	case reflect.Interface:
		v.Set(reflect.ValueOf(token.val))
	}
	return nil
}

func decodeList(l *lexer, val reflect.Value) error {
	v := indirect(val)
	if v.Kind() == reflect.Interface {
		i, err := consumeList(l)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(i))
		return nil
	}

	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		return fmt.Errorf("Cant store a []interface{} into %s", v.Type())
	}

	head := l.nextToken()
	if head.typ != listStartType {
		return fmt.Errorf("Can't decode list. Found: %s", head)
	}

	for i := 0; ; i++ {
		switch next := l.peekToken(); next.typ {
		case listEndType:
			l.nextToken() //consume end
			return nil
		case eofType:
			return errors.New("Unexpected EOF")
		case errorType:
			return l.nextToken() //consume the error
		}

		//grow it
		if i >= v.Cap() && v.IsValid() {
			newcap := v.Cap() + v.Cap()/2
			if newcap < 4 {
				newcap = 4
			}
			newv := reflect.MakeSlice(v.Type(), v.Len(), newcap)
			reflect.Copy(newv, v)
			v.Set(newv)
		}

		//reslice into cap (its a slice now since it had to have grown)
		if i >= v.Len() && v.IsValid() {
			v.SetLen(i + 1)
		}

		//decode a value into the index
		if err := decodeInto(l, v.Index(i)); err != nil {
			return err
		}
	}

	panic("unreachable")
}

func decodeDict(l *lexer, val reflect.Value) error {
	v := indirect(val)

	if v.Kind() == reflect.Interface {
		o, err := consumeDict(l)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(o))
		return nil
	}

	head := l.nextToken()
	if head.typ != dictStartType {
		return fmt.Errorf("Cant decode dict. Found: %s", head)
	}

	//check for correct type
	var (
		f       reflect.StructField
		mapElem reflect.Value
		isMap   bool
	)

	switch v.Kind() {
	case reflect.Map:
		t := v.Type()
		if t.Key() != reflectStringType {
			return fmt.Errorf("Can't store a map[string]interface{} into %s", v.Type())
		}
		if v.IsNil() {
			v.Set(reflect.MakeMap(t))
		}

		isMap = true
		mapElem = reflect.New(t.Elem()).Elem()
	case reflect.Struct:
	default:
		return fmt.Errorf("Can't store a map[string]interface{} into %s", v.Type())
	}

	for {
		var subv reflect.Value

		key := l.nextToken()
		switch key.typ {
		case dictEndType:
			return nil
		case eofType:
			return errors.New("Unexpected EOF")
		case errorType:
			return key
		}

		switch l.peekToken().typ {
		case eofType:
			return errors.New("Unexpected EOF")
		case dictEndType:
			return errors.New("Unexpected Dict End")
		case errorType:
			return l.nextToken() //consume the error
		}

		if isMap {
			mapElem.Set(reflect.Zero(v.Type().Elem()))
			subv = mapElem
		} else {
			var ok bool
			t := v.Type()
			if isValidTag(key.val) {
				for i := 0; i < v.NumField(); i++ {
					f = t.Field(i)
					tagName, _ := parseTag(f.Tag.Get("bencode"))
					if tagName == key.val && tagName != "-" {
						// If we have found a matching tag
						// that isn't '-'
						ok = true
						break
					}
				}
			}
			if !ok {
				f, ok = t.FieldByName(key.val)
			}
			if !ok {
				f, ok = t.FieldByNameFunc(matchName(key.val))
			}

			if ok {
				if f.PkgPath != "" {
					return fmt.Errorf("Can't store into unexported field: %s", f)
				}
				subv = v.FieldByIndex(f.Index)
			}
		}

		if !subv.IsValid() {
			//if it's invalid, grab but ignore the next value
			_, err := nextValue(l)
			if err != nil {
				return err
			}

			continue
		}

		//subv now contains what we load into
		if err := decodeInto(l, subv); err != nil {
			return err
		}

		if isMap {
			v.SetMapIndex(reflect.ValueOf(key.val), subv)
		}
	}

	panic("unreachable")
}
