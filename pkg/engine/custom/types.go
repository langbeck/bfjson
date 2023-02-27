package custom

import (
	"bytes"
	"embed"
	"text/template"
)

var (
	//go:embed templates/*.gotmpl
	templateFS embed.FS

	templates *template.Template
)

func init() {
	t, err := template.ParseFS(templateFS, "templates/*.gotmpl")
	if err != nil {
		panic(err)
	}

	templates = t
}

type StructInfo struct {
	Name                  string
	Type                  string
	ObjectDecoder         string
	ObjectPtrDecoder      string
	ObjectSliceDecoder    string
	ObjectSlicePtrDecoder string
	ObjectPool            string
	ObjectReleaser        string
	Fields                []*StructFieldInfo
}

type StructFieldInfo struct {
	Name     string
	NameJSON string
	TypeName string
	Default  *string

	IsUnmarshaler bool
	IsRawMessage  bool
	IsPointer     bool
	IsReleasable  bool

	ExtAllowSingle bool

	DecodeInfo
}

type DecodeInfo struct {
	DecoderRef string
	IsBasic    bool
	IsObject   bool
}

func (s *StructInfo) MarshalText() (text []byte, err error) {
	var buf bytes.Buffer
	err = templates.ExecuteTemplate(&buf, "object.gotmpl", *s)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
