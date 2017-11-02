package bencode

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestDecode(t *testing.T) {
	type testCase struct {
		in     string
		val    interface{}
		expect interface{}
		err    bool
	}

	type dT struct {
		X string
		Y int
		Z string `bencode:"zff"`
	}

	type Embedded struct {
		B string
	}

	type issue22 struct {
		X     string       `bencode:"x"`
		Time  myTimeType   `bencode:"t"`
		Foo   myBoolType   `bencode:"f"`
		Bar   myStringType `bencode:"b"`
		Slice mySliceType  `bencode:"s"`
		Y     string       `bencode:"y"`
	}

	type issue22WithErrorChild struct {
		Name  string           `bencode:"n"`
		Error errorMarshalType `bencode:"e"`
	}

	now := time.Now()

	var decodeCases = []testCase{
		//integers
		{`i5e`, new(int), int(5), false},
		{`i-10e`, new(int), int(-10), false},
		{`i8e`, new(uint), uint(8), false},
		{`i8e`, new(uint8), uint8(8), false},
		{`i8e`, new(uint16), uint16(8), false},
		{`i8e`, new(uint32), uint32(8), false},
		{`i8e`, new(uint64), uint64(8), false},
		{`i8e`, new(int), int(8), false},
		{`i8e`, new(int8), int8(8), false},
		{`i8e`, new(int16), int16(8), false},
		{`i8e`, new(int32), int32(8), false},
		{`i8e`, new(int64), int64(8), false},
		{`i0e`, new(*int), new(int), false},
		{`i-2e`, new(uint), nil, true},

		//bools
		{`i1e`, new(bool), true, false},
		{`i0e`, new(bool), false, false},
		{`i0e`, new(*bool), new(bool), false},
		{`i8e`, new(bool), true, false},

		//strings
		{`3:foo`, new(string), "foo", false},
		{`4:foob`, new(string), "foob", false},
		{`0:`, new(*string), new(string), false},
		{`6:short`, new(string), nil, true},

		//lists
		{`l3:foo3:bare`, new([]string), []string{"foo", "bar"}, false},
		{`li15ei20ee`, new([]int), []int{15, 20}, false},
		{`ld3:fooi0eed3:bari1eee`, new([]map[string]int), []map[string]int{
			{"foo": 0},
			{"bar": 1},
		}, false},

		//dicts

		{`d3:foo3:bar4:foob3:fooe`, new(map[string]string), map[string]string{
			"foo":  "bar",
			"foob": "foo",
		}, false},
		{`d1:X3:foo1:Yi10e3:zff3:bare`, new(dT), dT{"foo", 10, "bar"}, false},

		// encoding/json takes, if set, the tag as name and doesn't falls back to the
		// struct field's name.
		{`d1:X3:foo1:Yi10e1:Z3:bare`, new(dT), dT{"foo", 10, ""}, false},

		{`d1:X3:foo1:Yi10e1:h3:bare`, new(dT), dT{"foo", 10, ""}, false},
		{`d3:fooli0ei1ee3:barli2ei3eee`, new(map[string][]int), map[string][]int{
			"foo": []int{0, 1},
			"bar": []int{2, 3},
		}, false},
		{`de`, new(map[string]string), map[string]string{}, false},

		//into interfaces
		{`i5e`, new(interface{}), int64(5), false},
		{`li5ee`, new(interface{}), []interface{}{int64(5)}, false},
		{`5:hello`, new(interface{}), "hello", false},
		{`d5:helloi5ee`, new(interface{}), map[string]interface{}{"hello": int64(5)}, false},

		//into values whose type support the Unmarshaler interface
		{`1:y`, new(myTimeType), nil, true},
		{fmt.Sprintf("i%de", now.Unix()), new(myTimeType), myTimeType{time.Unix(now.Unix(), 0)}, false},
		{`1:y`, new(myBoolType), myBoolType(true), false},
		{`i42e`, new(myBoolType), nil, true},
		{`1:n`, new(myBoolType), myBoolType(false), false},
		{`1:n`, new(errorMarshalType), nil, true},
		{`li102ei111ei111ee`, new(myStringType), myStringType("foo"), false},
		{`i42e`, new(myStringType), nil, true},
		{`d1:ai1e3:foo3:bare`, new(mySliceType), mySliceType{"a", int64(1), "foo", "bar"}, false},
		{`i42e`, new(mySliceType), nil, true},

		//into values who have a child which type supports the Unmarshaler interface
		{
			fmt.Sprintf(`d1:x1:x1:ti%de1:f1:y1:b3:foo1:sd1:f3:foo1:ai42ee1:y1:ye`, now.Unix()),
			new(issue22),
			issue22{
				X:     "x",
				Time:  myTimeType{time.Unix(now.Unix(), 0)},
				Foo:   myBoolType(true),
				Bar:   myStringType("foo"),
				Slice: mySliceType{"a", int64(42), "f", "foo"},
				Y:     "y",
			},
			false,
		},
		{
			`d1:ei42e1:n3:fooe`,
			new(issue22WithErrorChild),
			nil,
			true,
		},

		//malformed
		{`i53:foo`, new(interface{}), nil, true},
		{`6:foo`, new(interface{}), nil, true},
		{`di5ei2ee`, new(interface{}), nil, true},
		{`d3:fooe`, new(interface{}), nil, true},
		{`l3:foo3:bar`, new(interface{}), nil, true},
		{`d-1:`, new(interface{}), nil, true},

		// embedded structs
		{`d1:A3:foo1:B3:bare`, new(struct {
			A string
			Embedded
		}), struct {
			A string
			Embedded
		}{"foo", Embedded{"bar"}}, false},
	}

	for i, tt := range decodeCases {
		err := DecodeString(tt.in, tt.val)
		if !tt.err && err != nil {
			t.Errorf("#%d: Unexpected err: %v", i, err)
			continue
		}
		if tt.err && err == nil {
			t.Errorf("#%d: Expected err is nil", i)
			continue
		}
		v := reflect.ValueOf(tt.val).Elem().Interface()
		if !reflect.DeepEqual(v, tt.expect) && !tt.err {
			t.Errorf("#%d: Val: %#v != %#v", i, v, tt.expect)
		}
	}
}

func TestRawDecode(t *testing.T) {
	type testCase struct {
		in     string
		expect []byte
		err    bool
	}

	var rawDecodeCases = []testCase{
		{`i5e`, []byte(`i5e`), false},
		{`5:hello`, []byte(`5:hello`), false},
		{`li5ei10e5:helloe`, []byte(`li5ei10e5:helloe`), false},
		{`llleee`, []byte(`llleee`), false},
		{`li5eli5eli5eeee`, []byte(`li5eli5eli5eeee`), false},
		{`d5:helloi5ee`, []byte(`d5:helloi5ee`), false},
	}

	for i, tt := range rawDecodeCases {
		var x RawMessage
		err := DecodeString(tt.in, &x)
		if !tt.err && err != nil {
			t.Errorf("#%d: Unexpected err: %v", i, err)
			continue
		}
		if tt.err && err == nil {
			t.Errorf("#%d: Expected err is nil", i)
			continue
		}
		if !reflect.DeepEqual(x, RawMessage(tt.expect)) && !tt.err {
			t.Errorf("#%d: Val: %#v != %#v", i, x, tt.expect)
		}
	}
}

type myStringType string

// UnmarshalBencode implements Unmarshaler.UnmarshalBencode
func (mst *myStringType) UnmarshalBencode(b []byte) error {
	var raw []byte
	err := DecodeBytes(b, &raw)
	if err != nil {
		return err
	}

	*mst = myStringType(raw)
	return nil
}

type mySliceType []interface{}

// UnmarshalBencode implements Unmarshaler.UnmarshalBencode
func (mst *mySliceType) UnmarshalBencode(b []byte) error {
	m := make(map[string]interface{})
	err := DecodeBytes(b, &m)
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	raw := make([]interface{}, 0, len(m)*2)
	for _, key := range keys {
		raw = append(raw, key, m[key])
	}

	*mst = mySliceType(raw)
	return nil
}

func TestNestedRawDecode(t *testing.T) {
	type testCase struct {
		in     string
		val    interface{}
		expect interface{}
		err    bool
	}

	type message struct {
		Key string
		Val int
		Raw RawMessage
	}

	var cases = []testCase{
		{`li5e5:hellod1:a1:beli5eee`, new([]RawMessage), []RawMessage{
			RawMessage(`i5e`),
			RawMessage(`5:hello`),
			RawMessage(`d1:a1:be`),
			RawMessage(`li5ee`),
		}, false},
		{`d1:a1:b1:c1:de`, new(map[string]RawMessage), map[string]RawMessage{
			"a": RawMessage(`1:b`),
			"c": RawMessage(`1:d`),
		}, false},
		{`d3:Key5:hello3:Rawldedei5e1:ae3:Vali10ee`, new(message), message{
			Key: "hello",
			Val: 10,
			Raw: RawMessage(`ldedei5e1:ae`),
		}, false},
	}

	for i, tt := range cases {
		err := DecodeString(tt.in, tt.val)
		if !tt.err && err != nil {
			t.Errorf("#%d: Unexpected err: %v", i, err)
			continue
		}
		if tt.err && err == nil {
			t.Errorf("#%d: Expected err is nil", i)
			continue
		}
		v := reflect.ValueOf(tt.val).Elem().Interface()
		if !reflect.DeepEqual(v, tt.expect) && !tt.err {
			t.Errorf("#%d: Val:\n%#v !=\n%#v", i, v, tt.expect)
		}
	}
}
