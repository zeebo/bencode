package bencode

import (
	"testing"
	"reflect"
)

type decodeTestCase struct {
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

var decodeCases = []decodeTestCase{
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
	{`i-2e`, new(uint), nil, true},

	//strings
	{`3:foo`, new(string), "foo", false},
	{`4:foob`, new(string), "foob", false},
	{`6:short`, new(string), nil, true},

	//lists
	{`l3:foo3:bare`, new([]string), []string{"foo", "bar"}, false},
	{`li15ei20ee`, new([]int), []int{15, 20}, false},

	//dicts
	{`d3:foo3:bar4:foob3:fooe`, new(map[string]string), map[string]string{
		"foo":  "bar",
		"foob": "foo",
	}, false},
	{`d1:X3:foo1:Yi10e3:zff3:bare`, new(dT), dT{"foo", 10, "bar"}, false},
	{`d1:X3:foo1:Yi10e1:Z3:bare`, new(dT), dT{"foo", 10, "bar"}, false},
	{`d1:X3:foo1:Yi10e1:h3:bare`, new(dT), dT{"foo", 10, ""}, false},

	//malformed
	{`i53:foo`, new(interface{}), nil, true},
	{`6:foo`, new(interface{}), nil, true},
	{`di5ei2ee`, new(interface{}), nil, true},
	{`d3:fooe`, new(interface{}), nil, true},
	{`l3:foo3:bar`, new(interface{}), nil, true},
}

func TestDecode(t *testing.T) {
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