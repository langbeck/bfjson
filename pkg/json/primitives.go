package json

import (
	"errors"
	"strconv"

	"github.com/langbeck/bfjson/pkg/json/tokens"
	"github.com/langbeck/bfjson/pkg/unsafe"
)

var ErrFormat = errors.New("format error")

var numberStart = [256]bool{
	'-': true,
	'0': true,
	'1': true,
	'2': true,
	'3': true,
	'4': true,
	'5': true,
	'6': true,
	'7': true,
	'8': true,
	'9': true,
}

func stringTokenToString(tok []byte) string {
	// return string(tok[1 : len(tok)-1])

	return unsafe.BytesToString(tok[1 : len(tok)-1])

	// t, _ := unquoteBytes(tok[1 : len(tok)-1])
	// return unsafe.BytesToString(t)

	// t, _ := unquoteBytes(tok[1 : len(tok)-1])
	// return string(t)
}

func parseInt(tok []byte) (int, error) {
	n, err := strconv.ParseInt(unsafe.BytesToString(tok), 10, 64)
	if err != nil {
		return 0, err
	}

	// TODO: do overflow check
	return int(n), nil
}

func (d *Decoder) DecodePtrInt(dst **int) error {
	tok, err := d.NextToken()
	if err != nil {
		return err
	}

	if tok[0] == tokens.Null {
		*dst = nil
		return nil
	}

	if !numberStart[tok[0]] {
		return ErrFormat
	}

	n, err := parseInt(tok)
	if err != nil {
		return err
	}

	*dst = &n
	return nil
}

func (d *Decoder) DecodeInt(dst *int) error {
	tok, err := d.NextToken()
	if err != nil {
		return err
	}

	if !numberStart[tok[0]] {
		return ErrFormat
	}

	n, err := parseInt(tok)
	if err != nil {
		return err
	}

	*dst = n
	return nil
}

func (d *Decoder) DecodeString(dst *string) error {
	tok, err := d.NextToken()
	if err != nil {
		return err
	}

	if tok[0] == tokens.Null {
		return nil
	}

	if tok[0] != tokens.String {
		return ErrFormat
	}

	*dst = stringTokenToString(tok)
	return nil
}

func (d *Decoder) DecodeFloat64(dst *float64) error {
	tok, err := d.NextToken()
	if err != nil {
		return err
	}

	if !numberStart[tok[0]] {
		return ErrFormat
	}

	value, err := strconv.ParseFloat(unsafe.BytesToString(tok), 64)
	if err != nil {
		return err
	}

	*dst = value
	return nil
}

func (d *Decoder) DecodeStringOrSlice(dst *[]string) error {
	return d.decodeSliceOfString(dst, true)
}

func (d *Decoder) DecodeSliceOfString(dst *[]string) error {
	return d.decodeSliceOfString(dst, false)
}

func (d *Decoder) decodeSliceOfString(dst *[]string, allowSingle bool) error {
	tok, err := d.NextToken()
	if err != nil {
		return err
	}

	if tok[0] == tokens.Null {
		*dst = nil
		return nil
	}

	if allowSingle && tok[0] == tokens.String {
		slice := []string{stringTokenToString(tok)}
		*dst = slice
		return nil
	}

	if tok[0] != tokens.ArrayStart {
		return ErrFormat
	}

	tok, err = d.NextToken()
	if err != nil {
		return err
	}

	if tok[0] == tokens.ArrayEnd {
		*dst = []string{}
		return nil
	}

	if tok[0] != tokens.String {
		return ErrFormat
	}

	slice := []string{stringTokenToString(tok)}
	for {
		tok, err := d.NextToken()
		if err != nil {
			return err
		}

		if tok[0] == tokens.ArrayEnd {
			break
		}

		if tok[0] != tokens.String {
			return ErrFormat
		}

		slice = append(slice, stringTokenToString(tok))
	}

	*dst = slice
	return nil
}

func (d *Decoder) DecodeIntOrSlice(dst *[]int) error {
	return d.decodeSliceOfInt(dst, true)
}

func (d *Decoder) DecodeSliceOfInt(dst *[]int) error {
	return d.decodeSliceOfInt(dst, false)
}

func (d *Decoder) decodeSliceOfInt(dst *[]int, allowSingle bool) error {
	tok, err := d.NextToken()
	if err != nil {
		return err
	}

	if tok[0] == tokens.Null {
		*dst = nil
		return nil
	}

	if allowSingle && numberStart[tok[0]] {
		n, err := parseInt(tok)
		if err != nil {
			return err
		}

		*dst = []int{n}
		return nil
	}

	if tok[0] != tokens.ArrayStart {
		return ErrFormat
	}

	tok, err = d.NextToken()
	if err != nil {
		return err
	}

	if tok[0] == tokens.ArrayEnd {
		*dst = []int{}
		return nil
	}

	if !numberStart[tok[0]] {
		return ErrFormat
	}

	n, err := parseInt(tok)
	if err != nil {
		return err
	}

	slice := []int{n}
	for {
		tok, err := d.NextToken()
		if err != nil {
			return err
		}

		if tok[0] == tokens.ArrayEnd {
			break
		}

		if !numberStart[tok[0]] {
			return ErrFormat
		}

		n, err := parseInt(tok)
		if err != nil {
			return err
		}

		slice = append(slice, n)
	}

	*dst = slice
	return nil
}
