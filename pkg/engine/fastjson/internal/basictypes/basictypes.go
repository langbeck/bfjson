package basictypes

import "go/types"

var (
	Error     = types.Universe.Lookup("error").Type()
	ByteSlice = types.NewSlice(types.Typ[types.Byte])

	JSONUnmarshaler = types.NewInterfaceType([]*types.Func{
		types.NewFunc(
			0,
			nil,
			"UnmarshalJSON",
			types.NewSignature(
				nil,
				types.NewTuple(types.NewParam(0, nil, "", ByteSlice)),
				types.NewTuple(types.NewParam(0, nil, "", Error)),
				false,
			),
		),
	}, nil).Complete()
)

func IsJSONUnmarshaler(typ types.Type) bool {
	return types.Implements(types.NewPointer(typ), JSONUnmarshaler)
}
