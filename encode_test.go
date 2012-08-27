package bencode

import "testing"

func TestEncode(t *testing.T) {
	type encodeTestCase struct {
		in  interface{}
		out string
		err bool
	}

	type eT struct {
		A string
		X string `bencode:"D"`
		Y string `bencode:"B"`
		Z string `bencode:"C"`
	}

	var encodeCases = []encodeTestCase{
		//integers
		{10, `i10e`, false},
		{-10, `i-10e`, false},
		{0, `i0e`, false},
		{int(10), `i10e`, false},
		{int8(10), `i10e`, false},
		{int16(10), `i10e`, false},
		{int32(10), `i10e`, false},
		{int64(10), `i10e`, false},
		{uint(10), `i10e`, false},
		{uint8(10), `i10e`, false},
		{uint16(10), `i10e`, false},
		{uint32(10), `i10e`, false},
		{uint64(10), `i10e`, false},

		//strings
		{"foo", `3:foo`, false},
		{"barbb", `5:barbb`, false},
		{"", `0:`, false},

		//lists
		{[]interface{}{"foo", 20}, `l3:fooi20ee`, false},
		{[]interface{}{90, 20}, `li90ei20ee`, false},
		{[]interface{}{[]interface{}{"foo", "bar"}, 20}, `ll3:foo3:barei20ee`, false},
		{[]map[string]int{
			{"a": 0, "b": 1},
			{"c": 2, "d": 3},
		}, `ld1:ai0e1:bi1eed1:ci2e1:di3eee`, false},

		//dicts
		{map[string]interface{}{
			"a": "foo",
			"c": "bar",
			"b": "tes",
		}, `d1:a3:foo1:b3:tes1:c3:bare`, false},
		{eT{"foo", "bar", "far", "boo"}, `d1:A3:foo1:B3:far1:C3:boo1:D3:bare`, false},
		{map[string][]int{
			"a": {0, 1},
			"b": {2, 3},
		}, `d1:ali0ei1ee1:bli2ei3eee`, false},
	}

	for i, tt := range encodeCases {
		data, err := EncodeString(tt.in)
		if !tt.err && err != nil {
			t.Errorf("#%d: Unexpected err: %v", i, err)
			continue
		}
		if tt.err && err == nil {
			t.Errorf("#%d: Expected err is nil", i)
			continue
		}
		if tt.out != data {
			t.Errorf("#%d: Val: %q != %q", i, data, tt.out)
		}
	}
}

func TestEncodeOmit(t *testing.T) {
	type encodeTestCase struct {
		in  interface{}
		out string
		err bool
	}

	type eT struct {
		A string `bencode:",omitempty"`
		B int    `bencode:",omitempty"`
		C *int   `bencode:",omitempty"`
	}

	var encodeCases = []encodeTestCase{
		{eT{}, `de`, false},
		{eT{A: "a"}, `d1:A1:ae`, false},
		{eT{B: 5}, `d1:Bi5ee`, false},
		{eT{C: new(int)}, `d1:Ci0ee`, false},
	}

	for i, tt := range encodeCases {
		data, err := EncodeString(tt.in)
		if !tt.err && err != nil {
			t.Errorf("#%d: Unexpected err: %v", i, err)
			continue
		}
		if tt.err && err == nil {
			t.Errorf("#%d: Expected err is nil", i)
			continue
		}
		if tt.out != data {
			t.Errorf("#%d: Val: %q != %q", i, data, tt.out)
		}
	}
}
