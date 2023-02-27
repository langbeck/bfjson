package json

import (
	"fmt"

	"github.com/langbeck/bfjson/pkg/json/internal/pkgjson"
	"github.com/langbeck/bfjson/pkg/json/tokens"
)

// Decoder must not be copied
type Decoder struct {
	pkgjson.Decoder
	boff int
	data []byte
}

func NewDecoder(data []byte) *Decoder {
	d := new(Decoder)
	d.Reset(data)
	return d
}

func (d *Decoder) Reset(data []byte) {
	d.Decoder.Reset(data)
	d.data = data
	d.boff = -1
}

func (d *Decoder) startBuffering() {
	if d.boff >= 0 {
		panic("already buffering")
	}

	d.boff = d.Offset()
}

func (d *Decoder) stopBuffering() []byte {
	if d.boff < 0 {
		panic("not buffering")
	}

	data := d.data[d.boff:d.Position()]
	d.boff = -1

	return data
}

func (d *Decoder) SkipAttribute() error {
	tok, err := d.NextToken()
	if err != nil {
		return err
	}

	if numberStart[tok[0]] {
		return nil
	}

	switch tok[0] {
	case tokens.Null, tokens.True, tokens.False, tokens.String:
		return nil

	case tokens.ObjectStart:
		return d.skipBallanced(tokens.ObjectStart, tokens.ObjectEnd, 1)

	case tokens.ArrayStart:
		return d.skipBallanced(tokens.ArrayStart, tokens.ArrayEnd, 1)

	default:
		return fmt.Errorf("SkipAttribute error: tok=%q", string(tok))
	}
}

func (d *Decoder) NextRawBytes() ([]byte, error) {
	tok, err := d.NextToken()
	if err != nil {
		return nil, err
	}

	d.startBuffering()
	if numberStart[tok[0]] {
		return d.stopBuffering(), nil
	}

	switch tok[0] {
	case tokens.Null, tokens.True, tokens.False, tokens.String:
		return d.stopBuffering(), nil

	case tokens.ObjectStart:
		err := d.skipBallanced(tokens.ObjectStart, tokens.ObjectEnd, 1)
		if err != nil {
			return nil, err
		}

		return d.stopBuffering(), nil

	case tokens.ArrayStart:
		err := d.skipBallanced(tokens.ArrayStart, tokens.ArrayEnd, 1)
		if err != nil {
			return nil, err
		}

		return d.stopBuffering(), nil

	default:
		return nil, fmt.Errorf("NextRawBytes error: tok=%q", string(tok))
	}
}

func (d *Decoder) skipBallanced(start, end byte, offset int) error {
	for {
		tok, err := d.NextToken()
		if err != nil {
			return err
		}

		switch tok[0] {
		case start:
			offset++

		case end:
			offset--
		}

		if offset == 0 {
			return nil
		}
	}
}
