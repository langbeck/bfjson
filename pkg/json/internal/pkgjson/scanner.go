package pkgjson

import (
	"github.com/langbeck/bfjson/pkg/json/tokens"
	"github.com/langbeck/bfjson/pkg/unsafe"
)

// Scanner implements a JSON scanner as defined in RFC 7159.
type Scanner struct {
	data []byte
	Off  int
	Pos  int
}

var whitespace = [256]bool{
	' ':  true,
	'\r': true,
	'\n': true,
	'\t': true,
}

var nextSimpleCase = [256]bool{
	tokens.ObjectStart: true,
	tokens.ObjectEnd:   true,
	tokens.Colon:       true,
	tokens.Comma:       true,
	tokens.ArrayStart:  true,
	tokens.ArrayEnd:    true,
}

// Next returns a []byte referencing the the next lexical token in the stream.
// The []byte is valid until Next is called again.
// If the stream is at its end, or an error has occured, Next returns a zero
// length []byte slice.
//
// A valid token begins with one of the following:
//
//	{ Object start
//	[ Array start
//	} Object end
//	] Array End
//	, Literal comma
//	: Literal colon
//	t JSON true
//	f JSON false
//	n JSON null
//	" A string, possibly containing backslash escaped entites.
//	-, 0-9 A number
func (s *Scanner) Next() []byte {
	s.Off = s.Pos

	data := s.data
	for pos := s.Pos; pos < len(data); pos++ {
		c := data[pos]

		// strip any leading whitespace.
		if whitespace[c] {
			continue
		}

		// simple case
		if nextSimpleCase[c] {
			s.Pos = pos + 1
			return data[pos:s.Pos]
		}

		s.Off = pos

		switch c {
		case tokens.True:
			if s.validateToken("true") == 0 {
				return nil
			}

		case tokens.False:
			if s.validateToken("false") == 0 {
				return nil
			}

		case tokens.Null:
			if s.validateToken("null") == 0 {
				return nil
			}

		case tokens.String:
			if s.parseString() < 2 {
				return nil
			}

		default:
			// ensure the number is correct.
			if s.parseNumber(c) == 0 {
				return nil
			}
		}

		return data[s.Off:s.Pos]
	}

	return nil
}

func (s *Scanner) validateToken(expected string) int {
	w := s.data[s.Off:]
	n := len(expected)
	if len(w) >= n {
		if unsafe.BytesToString(w[:n]) != expected {
			// doesn't match
			return 0
		}

		s.Pos = s.Off + n
		return n
	}

	// not enough data is left: eof
	return 0
}

func (s *Scanner) parseString() int {
	data := s.data
	for pos := s.Off + 1; pos < len(data); pos++ {
		switch data[pos] {
		case '"':
			// finished
			l := pos - s.Pos + 1
			s.Pos = pos + 1
			return l

		case '\\':
			pos++
		}
	}

	return 0
}

func (s *Scanner) parseNumber(c byte) int {
	const (
		begin = iota
		leadingzero
		anydigit1
		decimal
		anydigit2
		exponent
		expsign
		anydigit3
	)

	pos := 0
	// w := s.br.window(0)
	w := s.data[s.Off:]

	// int vs uint8 costs 10% on canada.json
	var state uint8 = begin

	// handle the case that the first character is a hyphen
	if c == '-' {
		pos++
		w = w[1:]
	}

	for {
		for _, elem := range w {
			switch state {
			case begin:
				if elem >= '1' && elem <= '9' {
					state = anydigit1
				} else if elem == '0' {
					state = leadingzero
				} else {
					// error
					return 0
				}
			case anydigit1:
				if elem >= '0' && elem <= '9' {
					// stay in this state
					break
				}
				fallthrough
			case leadingzero:
				if elem == '.' {
					state = decimal
					break
				}
				if elem == 'e' || elem == 'E' {
					state = exponent
					break
				}
				// finished
				s.Pos = s.Off + pos
				return pos
			case decimal:
				if elem >= '0' && elem <= '9' {
					state = anydigit2
				} else {
					// error
					return 0
				}
			case anydigit2:
				if elem >= '0' && elem <= '9' {
					break
				}
				if elem == 'e' || elem == 'E' {
					state = exponent
					break
				}
				// finished
				s.Pos = s.Off + pos
				return pos
			case exponent:
				if elem == '+' || elem == '-' {
					state = expsign
					break
				}
				fallthrough
			case expsign:
				if elem >= '0' && elem <= '9' {
					state = anydigit3
					break
				}
				// error
				return 0
			case anydigit3:
				if elem < '0' || elem > '9' {
					// finished
					s.Pos = s.Off + pos
					return pos
				}
			}
			pos++
		}

		// need more data from the pipe
		// end of the item. However, not necessarily an error. Make
		// sure we are in a state that allows ending the number.
		switch state {
		case leadingzero, anydigit1, anydigit2, anydigit3:
			s.Pos = s.Off + pos
			return pos
		default:
			// error otherwise, the number isn't complete.
			return 0
		}
	}
}
