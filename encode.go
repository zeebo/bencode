package bencode

import (
	"io"
	"os"
	"reflect"
	"fmt"
	"sort"
	"bytes"
)

type sortValues []reflect.Value

func (p sortValues) Len() int           { return len(p) }
func (p sortValues) Less(i, j int) bool { return p[i].String() < p[j].String() }
func (p sortValues) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type sortFields []reflect.StructField

func (p sortFields) Len() int           { return len(p) }
func (p sortFields) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p sortFields) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

func (e *Encoder) Encode(val interface{}) os.Error {
	return encodeValue(e.w, reflect.ValueOf(val))
}

func EncodeString(val interface{}) (string, os.Error) {
	buf := new(bytes.Buffer)
	e := NewEncoder(buf)
	if err := e.Encode(val); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func encodeValue(w io.Writer, val reflect.Value) os.Error {
	//inspect the val to check
	v := indirect(val)

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		_, err := fmt.Fprintf(w, "i%de", v.Int())
		return err

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		_, err := fmt.Fprintf(w, "i%de", v.Uint())
		return err

	case reflect.String:
		_, err := fmt.Fprintf(w, "%d:%s", len(v.String()), v.String())
		return err

	case reflect.Slice, reflect.Array:
		if _, err := fmt.Fprint(w, "l"); err != nil {
			return err
		}
		for i := 0; i < v.Len(); i++ {
			if err := encodeValue(w, v.Index(i)); err != nil {
				return err
			}
		}

		_, err := fmt.Fprint(w, "e")
		return err
		

	case reflect.Map:
		if _, err := fmt.Fprint(w, "d"); err != nil {
			return err
		}
		var (
			keys sortValues = v.MapKeys()
			mval reflect.Value
		)
		sort.Sort(keys)
		for i := range keys {
			if err := encodeValue(w, keys[i]); err != nil {
				return err
			}
			mval = v.MapIndex(keys[i])
			if err := encodeValue(w, mval); err != nil {
				return err
			}
		}
		_, err := fmt.Fprint(w, "e");
		return err

	case reflect.Struct:
		t := v.Type()
		if _, err := fmt.Fprint(w, "d"); err != nil {
			return err
		}
		//put keys into keys
		var (
			keys = make(sortFields, t.NumField())
			mval reflect.Value
		)
		for i := range keys {
			keys[i] = t.Field(i)
		}
		sort.Sort(keys)
		for _, key := range keys {
			if err := encodeValue(w, reflect.ValueOf(key.Name)); err != nil {
				return err
			}
			mval = v.FieldByIndex(key.Index)
			if err := encodeValue(w, mval); err != nil {
				return err
			}
		}
		_, err := fmt.Fprint(w, "e")
		return err
	}

	return fmt.Errorf("Can't encode type: %s", v.Type())
}
