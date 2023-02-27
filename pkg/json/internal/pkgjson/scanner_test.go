package pkgjson

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

type SmallReader struct {
	r io.Reader
	n int
}

func (sm *SmallReader) next() int {
	sm.n = (sm.n + 3) % 5
	if sm.n < 1 {
		sm.n++
	}
	return sm.n
}

func (sm *SmallReader) Read(buf []byte) (int, error) {
	return sm.r.Read(buf[:min(sm.next(), len(buf))])
}

func TestScannerNext(t *testing.T) {
	tests := []struct {
		in     string
		tokens []string
	}{
		{in: `""`, tokens: []string{`""`}},
		{in: `"a"`, tokens: []string{`"a"`}},
		{in: ` "a" `, tokens: []string{`"a"`}},
		{in: `"\""`, tokens: []string{`"\""`}},
		{in: `{}`, tokens: []string{`{`, `}`}},
		{in: `[]`, tokens: []string{`[`, `]`}},
		{in: `[{}, {}]`, tokens: []string{`[`, `{`, `}`, `,`, `{`, `}`, `]`}},
		{in: `{"a": 0}`, tokens: []string{`{`, `"a"`, `:`, `0`, `}`}},
		{in: `{"a": []}`, tokens: []string{`{`, `"a"`, `:`, `[`, `]`, `}`}},
		{in: `[10]`, tokens: []string{`[`, `10`, `]`}},
		{in: `{"x": "va\\\\ue", "y": "value y"}`, tokens: []string{
			`{`, `"x"`, `:`, `"va\\\\ue"`, `,`, `"y"`, `:`, `"value y"`, `}`,
		}},
		{in: `1`, tokens: []string{`1`}},
		{in: `{"c": null}`, tokens: []string{"{", `"c"`, ":", "null", "}"}},
		{in: `-1234567.8e+90`, tokens: []string{`-1234567.8e+90`}},
		{in: `[{"a": 1,"b": 123.456, "c": null, "d": [1, -2, "three", true, false, ""]}]`,
			tokens: []string{`[`,
				`{`,
				`"a"`, `:`, `1`, `,`,
				`"b"`, `:`, `123.456`, `,`,
				`"c"`, `:`, `null`, `,`,
				`"d"`, `:`, `[`,
				`1`, `,`, `-2`, `,`, `"three"`, `,`, `true`, `,`, `false`, `,`, `""`,
				`]`,
				`}`,
				`]`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			scanner := &Scanner{data: []byte(tc.in)}
			for n, want := range tc.tokens {
				got := scanner.Next()
				if string(got) != want {
					t.Fatalf("%+v: expected: %q, got: %q (%x) %v", n+1, want, string(got), got, got == nil)
				}
			}
			last := scanner.Next()
			if len(last) > 0 {
				t.Fatalf("expected: %q, got: %q", "", string(last))
			}
			// if err := scanner.Error(); err != io.EOF {
			// 	t.Fatalf("expected: %v, got: %v", io.EOF, err)
			// }
		})
	}
}

func TestParseString(t *testing.T) {
	testParseString(t, `""`, `""`)
	testParseString(t, `"" `, `""`)
	testParseString(t, `"\""`, `"\""`)
	testParseString(t, `"\\\\\\\\\6"`, `"\\\\\\\\\6"`)
	testParseString(t, `"\6"`, `"\6"`)
}

func testParseString(t *testing.T, json, want string) {
	t.Helper()
	scanner := &Scanner{data: []byte(json)}
	got := scanner.Next()
	if string(got) != want {
		t.Fatalf("expected: %q, got: %q", want, got)
	}
}

func TestParseNumber(t *testing.T) {
	testParseNumber(t, `1`)
	// testParseNumber(t, `0000001`)
	testParseNumber(t, `12.0004`)
	testParseNumber(t, `1.7734`)
	testParseNumber(t, `15`)
	testParseNumber(t, `-42`)
	testParseNumber(t, `-1.7734`)
	testParseNumber(t, `1.0e+28`)
	testParseNumber(t, `-1.0e+28`)
	testParseNumber(t, `1.0e-28`)
	testParseNumber(t, `-1.0e-28`)
	testParseNumber(t, `-18.3872`)
	testParseNumber(t, `-2.1`)
	testParseNumber(t, `-1234567.891011121314`)
}

func testParseNumber(t *testing.T, tc string) {
	t.Helper()
	scanner := &Scanner{data: []byte(tc)}
	got := scanner.Next()
	if string(got) != tc {
		t.Fatalf("expected: %q, got: %q", tc, got)
	}
}

func BenchmarkParseNumber(b *testing.B) {
	tests := []string{
		`1`,
		`12.0004`,
		`1.7734`,
		`15`,
		`-42`,
		`-1.7734`,
		`1.0e+28`,
		`-1.0e+28`,
		`1.0e-28`,
		`-1.0e-28`,
		`-18.3872`,
		`-2.1`,
		`-1234567.891011121314`,
	}

	for _, tc := range tests {
		b.Run(tc, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				scanner := &Scanner{
					data: []byte(tc),
				}
				n := scanner.parseNumber(scanner.data[0])
				if n != len(tc) {
					b.Fatalf("failed")
				}
			}
		})
	}
}

func TestScanner(t *testing.T) {
	for _, tc := range inputs {
		r := fixture(t, tc.path)
		t.Run(tc.path, func(t *testing.T) {
			sc := &Scanner{
				data: r,
			}
			n := 0
			for len(sc.Next()) > 0 {
				n++
			}
			if n != tc.alltokens {
				t.Fatalf("expected %v tokens, got %v", tc.alltokens, n)
			}
		})
	}
}

var inputs = []struct {
	path       string
	tokens     int // decoded tokens
	alltokens  int // raw tokens, includes : and ,
	whitespace int // number of whitespace chars
}{
	// from https://github.com/miloyip/nativejson-benchmark
	{"canada", 223236, 334373, 33},
	{"citm_catalog", 85035, 135990, 1227563},
	{"twitter", 29573, 55263, 167931},
	{"code", 217707, 396293, 3},

	// from https://raw.githubusercontent.com/mailru/easyjson/master/benchmark/example.json
	{"example", 710, 1297, 4246},

	// from https://github.com/ultrajson/ultrajson/blob/master/tests/sample.json
	{"sample", 5276, 8677, 518549},
}

// fuxture returns a *bytes.Reader for the contents of path.
func fixture(tb testing.TB, path string) []byte {
	data, err := os.ReadFile(filepath.Join("testdata", path+".json"))
	check(tb, err)
	return data
}

func check(tb testing.TB, err error) {
	if err != nil {
		tb.Helper()
		tb.Fatal(err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
