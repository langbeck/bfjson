package json

import (
	"io"
	"testing"
)

var tests = []struct {
	json          string
	before, after []string
	wantRaw       string
}{
	{
		json:    `[[[[[[{"true":true}]]]]]]`,
		before:  []string{`[`, `[`, `[`, `[`, `[`, `[`},
		after:   []string{`]`, `]`, `]`, `]`, `]`, `]`},
		wantRaw: `{"true":true}`,
	},
	{
		json:    `{"a": 0}`,
		before:  []string{},
		after:   []string{},
		wantRaw: `{"a": 0}`,
	},
	{
		json:    `[{}, {}]`,
		before:  []string{`[`},
		after:   []string{`{`, `}`, `]`},
		wantRaw: `{}`,
	},
	{
		json:    `[{ }, {}]`,
		before:  []string{`[`},
		after:   []string{`{`, `}`, `]`},
		wantRaw: `{ }`,
	},
	{
		json:    `{"a": []}`,
		before:  []string{},
		after:   []string{},
		wantRaw: `{"a": []}`,
	},
	{
		json:    `[{}]`,
		before:  []string{`[`},
		after:   []string{`]`},
		wantRaw: `{}`,
	},
	{
		json:    `{ }`,
		before:  []string{},
		after:   []string{},
		wantRaw: `{ }`,
	},
	{
		json:    `""`,
		before:  []string{},
		after:   []string{},
		wantRaw: `""`,
	},
	{
		json:    `[{"a": [{}]}]`,
		before:  []string{`[`},
		after:   []string{`]`},
		wantRaw: `{"a": [{}]}`,
	},
	{
		json:    `[10]`,
		before:  []string{`[`},
		after:   []string{`]`},
		wantRaw: `10`,
	},
	{
		json:    `1`,
		before:  []string{},
		after:   []string{},
		wantRaw: `1`,
	},
	{
		json:    `[{"a": 1,"b": 123.456, "c": null, "d": [1, -2, "three", true, false, ""]}]`,
		before:  []string{`[`},
		after:   []string{`]`},
		wantRaw: `{"a": 1,"b": 123.456, "c": null, "d": [1, -2, "three", true, false, ""]}`,
	},
}

func TestNextRawBytes(t *testing.T) {
	for _, test := range tests {
		dec := NewDecoder([]byte(test.json))
		for _, wantBefore := range test.before {
			got, err := dec.NextToken()
			if err != nil {
				t.Fatal(err)
			}

			if string(got) != wantBefore {
				t.Fatalf("before: want %s and got %s", wantBefore, string(got))
			}
		}

		gotRaw, err := dec.NextRawBytes()
		if err != nil {
			t.Fatal(err)
		}

		if string(gotRaw) != test.wantRaw {
			t.Fatalf("raw: want %s and got %s", test.wantRaw, string(gotRaw))
		}

		for _, wantAfter := range test.after {
			got, err := dec.NextToken()
			if err != nil {
				t.Fatal(err)
			}

			if string(got) != wantAfter {
				t.Fatalf("after: want %s and got %s", wantAfter, string(got))
			}
		}

		_, err = dec.NextToken()
		if err != io.EOF {
			t.Fatalf("err: want io.EOF, got %v", err)
		}
	}
}

func TestSkipAttribute(t *testing.T) {
	for _, test := range tests {
		dec := NewDecoder([]byte(test.json))
		for _, wantBefore := range test.before {
			got, err := dec.NextToken()
			if err != nil {
				t.Fatal(err)
			}

			if string(got) != wantBefore {
				t.Fatalf("before: want %s and got %s", wantBefore, string(got))
			}
		}

		err := dec.SkipAttribute()
		if err != nil {
			t.Fatal(err)
		}

		for _, wantAfter := range test.after {
			got, err := dec.NextToken()
			if err != nil {
				t.Fatal(err)
			}

			if string(got) != wantAfter {
				t.Fatalf("after: want %s and got %s", wantAfter, string(got))
			}
		}

		_, err = dec.NextToken()
		if err != io.EOF {
			t.Fatalf("err: want io.EOF, got %v", err)
		}
	}
}
