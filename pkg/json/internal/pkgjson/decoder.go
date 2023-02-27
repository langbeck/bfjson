// Package json decodes JSON.
package pkgjson

import (
	"fmt"
	"io"

	. "github.com/langbeck/bfjson/pkg/json/tokens"
)

// A Decoder decodes JSON values.
type Decoder struct {
	scanner Scanner
	state   func(*Decoder) ([]byte, error)
	stack
}

func NewDecoder(data []byte) *Decoder {
	d := new(Decoder)
	d.Reset(data)
	return d
}

func (d *Decoder) Reset(data []byte) {
	*d = Decoder{
		scanner: Scanner{data: data},
		state:   (*Decoder).stateValue,
	}
}

func (d *Decoder) Offset() int {
	return d.scanner.Off
}

func (d *Decoder) Position() int {
	return d.scanner.Pos
}

type stack []bool

func (s *stack) push(v bool) {
	*s = append(*s, v)
}

func (s *stack) pop() bool {
	*s = (*s)[:len(*s)-1]
	if len(*s) == 0 {
		return false
	}
	return (*s)[len(*s)-1]
}

func (s *stack) len() int { return len(*s) }

// NextToken returns a []byte referencing the next logical token in the stream.
// The []byte is valid until Token is called again.
// At the end of the input stream, Token returns nil, io.EOF.
//
// Token guarantees that the delimiters [ ] { } it returns are properly nested
// and matched: if Token encounters an unexpected delimiter in the input, it
// will return an error.
//
// A valid token begins with one of the following:
//
//	{ Object start
//	[ Array start
//	} Object end
//	] Array End
//	t JSON true
//	f JSON false
//	n JSON null
//	" A string, possibly containing backslash escaped entites.
//	-, 0-9 A number
//
// Commas and colons are elided.
func (d *Decoder) NextToken() ([]byte, error) {
	return d.state(d)
}

func (d *Decoder) stateObjectString() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '}':
		inObj := d.pop()
		switch {
		case d.len() == 0:
			d.state = (*Decoder).stateEnd
		case inObj:
			d.state = (*Decoder).stateObjectComma
		case !inObj:
			d.state = (*Decoder).stateArrayComma
		}
		return tok, nil
	case '"':
		d.state = (*Decoder).stateObjectColon
		return tok, nil
	default:
		return nil, fmt.Errorf("stateObjectString: missing string key")
	}
}

func (d *Decoder) stateObjectColon() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case Colon:
		d.state = (*Decoder).stateObjectValue
		return d.NextToken()
	default:
		return tok, fmt.Errorf("stateObjectColon: expecting colon")
	}
}

func (d *Decoder) stateObjectValue() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '{':
		d.state = (*Decoder).stateObjectString
		d.push(true)
		return tok, nil
	case '[':
		d.state = (*Decoder).stateArrayValue
		d.push(false)
		return tok, nil
	default:
		d.state = (*Decoder).stateObjectComma
		return tok, nil
	}
}

func (d *Decoder) stateObjectComma() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '}':
		inObj := d.pop()
		switch {
		case d.len() == 0:
			d.state = (*Decoder).stateEnd
		case inObj:
			d.state = (*Decoder).stateObjectComma
		case !inObj:
			d.state = (*Decoder).stateArrayComma
		}
		return tok, nil
	case Comma:
		d.state = (*Decoder).stateObjectString
		return d.NextToken()
	default:
		return tok, fmt.Errorf("stateObjectComma: expecting comma")
	}
}

func (d *Decoder) stateArrayValue() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '{':
		d.state = (*Decoder).stateObjectString
		d.push(true)
		return tok, nil
	case '[':
		d.state = (*Decoder).stateArrayValue
		d.push(false)
		return tok, nil
	case ']':
		inObj := d.pop()
		switch {
		case d.len() == 0:
			d.state = (*Decoder).stateEnd
		case inObj:
			d.state = (*Decoder).stateObjectComma
		case !inObj:
			d.state = (*Decoder).stateArrayComma
		}
		return tok, nil
	case ',':
		return nil, fmt.Errorf("stateArrayValue: unexpected comma")
	default:
		d.state = (*Decoder).stateArrayComma
		return tok, nil
	}
}

func (d *Decoder) stateArrayComma() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case ']':
		inObj := d.pop()
		switch {
		case d.len() == 0:
			d.state = (*Decoder).stateEnd
		case inObj:
			d.state = (*Decoder).stateObjectComma
		case !inObj:
			d.state = (*Decoder).stateArrayComma
		}
		return tok, nil
	case Comma:
		d.state = (*Decoder).stateArrayValue
		return d.NextToken()
	default:
		return nil, fmt.Errorf("stateArrayComma: expected comma, %v", d.stack)
	}
}

func (d *Decoder) stateValue() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '{':
		d.state = (*Decoder).stateObjectString
		d.push(true)
		return tok, nil
	case '[':
		d.state = (*Decoder).stateArrayValue
		d.push(false)
		return tok, nil
	case ',':
		return nil, fmt.Errorf("stateValue: unexpected comma")
	default:
		d.state = (*Decoder).stateEnd
		return tok, nil
	}
}

func (d *Decoder) stateEnd() ([]byte, error) { return nil, io.EOF }
