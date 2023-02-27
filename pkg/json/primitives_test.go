package json

import (
	"math"
	"reflect"
	"testing"
)

func TestDecodeInt(t *testing.T) {
	tests := []struct {
		json      string
		value     int
		shouldErr bool
	}{
		{json: `1`, value: 1, shouldErr: false},
		{json: `10`, value: 10, shouldErr: false},
		// { json: `01`, value: 1, err:  false }, TODO: fix leading 0 interprt.
		{json: `-1`, value: -1, shouldErr: false},
		{json: `127`, value: math.MaxInt8, shouldErr: false},
		{json: `-128`, value: math.MinInt8, shouldErr: false},
		{json: `32767`, value: math.MaxInt16, shouldErr: false},
		{json: `-32768`, value: math.MinInt16, shouldErr: false},
		{json: `2147483647`, value: math.MaxInt32, shouldErr: false},
		{json: `-2147483648`, value: math.MinInt32, shouldErr: false},
		{json: `9223372036854775807`, value: math.MaxInt64, shouldErr: false},
		{json: `-9223372036854775808`, value: math.MinInt64, shouldErr: false},
		{json: `1.0`, value: 0, shouldErr: true},
		{json: `[1]`, value: 0, shouldErr: true},
		{json: `null`, value: 0, shouldErr: true},
		// {json: `a`, value: 0, shouldErr: false}, // TODO: got unexpected EOF
		{json: `{}`, value: 0, shouldErr: true},
		{json: `[{}]`, value: 0, shouldErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.json, func(t *testing.T) {
			d := NewDecoder([]byte(tt.json))
			var got int
			err := d.DecodeInt(&got)

			gotErr := err != nil
			if tt.shouldErr != gotErr {
				t.Errorf("err: didn't want an error but got %v", err)
			}
			if got != tt.value {
				t.Errorf("value: want %d got %d", tt.value, got)
			}
		})
	}
}

func TestDecodeFloat64(t *testing.T) {
	tests := []struct {
		json      string
		value     float64
		shouldErr bool
	}{
		{json: `1.0`, value: 1.0, shouldErr: false},
		{json: `10.1`, value: 10.1, shouldErr: false},
		// { json: `01.0`, value: 1.0, err:  false }, TODO: fix leading 0 interprt.
		{json: `-1`, value: -1, shouldErr: false},
		{json: `-1.0`, value: -1.0, shouldErr: false},
		{json: `0`, value: 0, shouldErr: false},
		{json: `1.797693134862315708145274237317043567981e+308`, value: math.MaxFloat64, shouldErr: false},
		{json: `4.940656458412465441765687928682213723651e-324`, value: math.SmallestNonzeroFloat64, shouldErr: false},
		{json: `[1]`, value: 0, shouldErr: true},
		{json: `null`, value: 0, shouldErr: true},
		// {json: `a`, value: 0, err: true}, // TODO: got unexpected EOF
		{json: `{}`, value: 0, shouldErr: true},
		{json: `{ }`, value: 0, shouldErr: true},
		{json: `[{}]`, value: 0, shouldErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.json, func(t *testing.T) {
			d := NewDecoder([]byte(tt.json))
			var got float64
			err := d.DecodeFloat64(&got)

			gotErr := err != nil
			if tt.shouldErr != gotErr {
				t.Errorf("err: didn't want an error but got %v", err)
			}
			if got != tt.value {
				t.Errorf("value: want %f got %f", tt.value, got)
			}
		})
	}
}

func TestDecodeString(t *testing.T) {
	tests := []struct {
		json      string
		value     string
		shouldErr bool
	}{
		{json: `"a"`, value: "a", shouldErr: false},
		{json: `"a,a"`, value: "a,a", shouldErr: false},
		{json: `"aba"`, value: "aba", shouldErr: false},
		{json: `"1"`, value: "1", shouldErr: false},
		{json: `"1.0"`, value: "1.0", shouldErr: false},
		{json: `"1.0a"`, value: "1.0a", shouldErr: false},
		{json: `"[1.0a]"`, value: "[1.0a]", shouldErr: false},
		{json: `"{a}"`, value: "{a}", shouldErr: false},
		{json: ``, value: "", shouldErr: true}, // TODO: should this be an error or empty string?
		{json: `[]`, value: "", shouldErr: true},
		{json: `["a"]`, value: "", shouldErr: true},
		{json: `{"a"}`, value: "", shouldErr: true},
		// {json: `null`, value: "", shouldErr: false}, // TODO: should this be format error?
	}
	for _, tt := range tests {
		t.Run(tt.json, func(t *testing.T) {
			d := NewDecoder([]byte(tt.json))
			var got string
			err := d.DecodeString(&got)

			gotErr := err != nil
			if tt.shouldErr != gotErr {
				t.Errorf("err: didn't want an error but got %v", err)
			}
			if got != tt.value {
				t.Errorf("value: want %s got %s", tt.value, got)
			}
		})
	}
}

func TestDecodePtrInt(t *testing.T) {
	tests := []struct {
		json      string
		value     int
		shouldErr bool
	}{
		{json: `1`, value: 1, shouldErr: false},
		{json: `10`, value: 10, shouldErr: false},
		// { json: `01`, value: 1, err:  false }, TODO: fix leading 0 interprt.
		{json: `-1`, value: -1, shouldErr: false},
		{json: `127`, value: math.MaxInt8, shouldErr: false},
		{json: `-128`, value: math.MinInt8, shouldErr: false},
		{json: `32767`, value: math.MaxInt16, shouldErr: false},
		{json: `-32768`, value: math.MinInt16, shouldErr: false},
		{json: `2147483647`, value: math.MaxInt32, shouldErr: false},
		{json: `-2147483648`, value: math.MinInt32, shouldErr: false},
		{json: `9223372036854775807`, value: math.MaxInt64, shouldErr: false},
		{json: `-9223372036854775808`, value: math.MinInt64, shouldErr: false},
		{json: `1.0`, value: 0, shouldErr: true},
		{json: `[1]`, value: 0, shouldErr: true},
		// {json: `a`, value: 0, err: true}, // TODO: got unexpected EOF
		{json: `{}`, value: 0, shouldErr: true},
		{json: `[{}]`, value: 0, shouldErr: true},
		// {json: `null`, value: 0, shouldErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.json, func(t *testing.T) {
			d := NewDecoder([]byte(tt.json))
			var got *int
			err := d.DecodePtrInt(&got)

			gotErr := err != nil
			if tt.shouldErr != gotErr {
				t.Fatalf("err: didn't want an error but got %v", err)
			}
			if !gotErr && *got != tt.value {
				t.Errorf("value: want %d got %d", tt.value, *got)
			}
		})
	}
}

func TestDecodeSliceOfString(t *testing.T) {
	tests := []struct {
		json      string
		value     []string
		shouldErr bool
	}{
		{json: `["a"]`, value: []string{"a"}, shouldErr: false},
		{json: `["a,a"]`, value: []string{"a,a"}, shouldErr: false},
		{json: `["aba"]`, value: []string{"aba"}, shouldErr: false},
		{json: `["1"]`, value: []string{"1"}, shouldErr: false},
		{json: `["1.0"]`, value: []string{"1.0"}, shouldErr: false},
		{json: `["1.0a"]`, value: []string{"1.0a"}, shouldErr: false},
		{json: `["[1.0a]"]`, value: []string{"[1.0a]"}, shouldErr: false},
		{json: `["{a}"]`, value: []string{"{a}"}, shouldErr: false},
		{json: `[]`, value: []string{}, shouldErr: false},
		{json: `"a"`, value: []string{}, shouldErr: true},
		{json: `"a,a"`, value: []string{}, shouldErr: true},
		{json: ``, value: []string{}, shouldErr: true}, // TODO: should this be an error or empty slice?
		{json: `{"a"}`, value: []string{}, shouldErr: true},
		// {json: `null`, value: []string{}, shouldErr: false}, // TODO: should this be format error?
	}
	for _, tt := range tests {
		t.Run(tt.json, func(t *testing.T) {
			d := NewDecoder([]byte(tt.json))
			var got []string
			err := d.DecodeSliceOfString(&got)
			gotErr := err != nil
			if tt.shouldErr != gotErr {
				t.Fatalf("err: didn't want an error but got %v", err)
			}
			if !gotErr && !reflect.DeepEqual(got, tt.value) {
				t.Errorf("value: want %s got %s", tt.value, got)
			}
		})
	}
}

func TestDecodeSliceOfInt(t *testing.T) {
	tests := []struct {
		json      string
		value     []int
		shouldErr bool
	}{
		{json: `[1]`, value: []int{1}, shouldErr: false},
		{json: `[10]`, value: []int{10}, shouldErr: false},
		// { j[son: `01]`, value: []int{1}, err:  false }, TODO: fix leading 0 interprt.
		{json: `[-1]`, value: []int{-1}, shouldErr: false},
		{json: `[127]`, value: []int{math.MaxInt8}, shouldErr: false},
		{json: `[-128]`, value: []int{math.MinInt8}, shouldErr: false},
		{json: `[32767]`, value: []int{math.MaxInt16}, shouldErr: false},
		{json: `[-32768]`, value: []int{math.MinInt16}, shouldErr: false},
		{json: `[2147483647]`, value: []int{math.MaxInt32}, shouldErr: false},
		{json: `[-2147483648]`, value: []int{math.MinInt32}, shouldErr: false},
		{json: `[9223372036854775807]`, value: []int{math.MaxInt64}, shouldErr: false},
		{json: `[-9223372036854775808]`, value: []int{math.MinInt64}, shouldErr: false},
		{json: `[1]`, value: []int{1}, shouldErr: false},
		{json: `[1.0]`, value: []int{}, shouldErr: true},
		{json: `[null]`, value: []int{}, shouldErr: true},
		{json: `[a]`, value: []int{}, shouldErr: true},
		{json: `[{}]`, value: []int{}, shouldErr: true},
		{json: `[[{}]]`, value: []int{}, shouldErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.json, func(t *testing.T) {
			d := NewDecoder([]byte(tt.json))
			var got []int
			err := d.DecodeSliceOfInt(&got)

			gotErr := err != nil
			if tt.shouldErr != gotErr {
				t.Fatalf("err: didn't want an error but got %v", err)
			}
			if !gotErr && !reflect.DeepEqual(got, tt.value) {
				t.Errorf("value: want %d got %d", tt.value, got)
			}
		})
	}
}
